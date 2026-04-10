package webui

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"filearchiver/internal/archiver"
)

type Server struct {
	ConfigPath string
}

var (
	indexETag    = buildETag(indexHTML)
	settingsETag = buildETag(settingsHTML)
)

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/settings", s.handleSettings)
	mux.HandleFunc("/api/config", s.handleConfig)
	mux.HandleFunc("/api/archive", s.handleArchive)
	mux.HandleFunc("/api/ai/test", s.handleAITest)
	mux.HandleFunc("/api/path/resolve", s.handleResolvePath)
	return gzipMiddleware(mux)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	serveEmbeddedHTML(w, r, indexHTML, indexETag)
}

func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	serveEmbeddedHTML(w, r, settingsHTML, settingsETag)
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		cfg, err := archiver.LoadConfig(s.ConfigPath)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, cfg)
	case http.MethodPost:
		var req archiver.Config
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "配置格式无效")
			return
		}
		if err := archiver.SaveConfig(s.ConfigPath, &req); err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "配置已保存"})
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleArchive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cfg, err := archiver.LoadConfig(s.ConfigPath)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	mr, err := r.MultipartReader()
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "上传请求无效")
		return
	}

	type itemResult struct {
		Filename        string `json:"filename"`
		RenamedFilename string `json:"renamedFilename,omitempty"`
		Category        string `json:"category,omitempty"`
		CategoryAbbr    string `json:"categoryAbbr,omitempty"`
		DestinationPath string `json:"destinationPath,omitempty"`
		Status          string `json:"status"`
		Error           string `json:"error,omitempty"`
	}

	var (
		category     string
		useAI        bool
		template     string
		results      []itemResult
		successCount int
		totalFiles   int
	)
	versionCache := archiver.NewVersionCache()

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "读取上传数据失败")
			return
		}

		name := part.FormName()
		switch name {
		case "category":
			category = strings.TrimSpace(readSmallFormValue(part, 4<<10))
		case "useAI":
			useAI = strings.EqualFold(strings.TrimSpace(readSmallFormValue(part, 16)), "true")
		case "template":
			template = strings.TrimSpace(readSmallFormValue(part, 8<<10))
		case "file":
			totalFiles++
			safeName := filepath.Base(part.FileName())
			if safeName == "." || safeName == string(filepath.Separator) || safeName == "" {
				safeName = "uploaded-file"
			}

			chosenTemplate := template
			if chosenTemplate == "" {
				chosenTemplate = cfg.TemplatePrefix
			}

			if useAI && !aiConfigured(cfg.AI) {
				_, _ = io.Copy(io.Discard, part)
				results = append(results, itemResult{Filename: safeName, Status: "failed", Error: "AI 配置不完整，请先配置 url/apiKey/modelName"})
				_ = part.Close()
				continue
			}

			chosenCategory := category
			if useAI {
				chosenCategory, err = classifyCategoryByFilename(r.Context(), safeName, cfg)
				if err != nil {
					_, _ = io.Copy(io.Discard, part)
					results = append(results, itemResult{Filename: safeName, Status: "failed", Error: err.Error()})
					_ = part.Close()
					continue
				}
			}
			if !useAI && strings.TrimSpace(chosenCategory) == "" {
				_, _ = io.Copy(io.Discard, part)
				results = append(results, itemResult{Filename: safeName, Status: "failed", Error: "分类不能为空"})
				_ = part.Close()
				continue
			}

			archiveResult, err := archiver.RunReaderWithConfigResult(part, safeName, chosenCategory, chosenTemplate, cfg, versionCache)
			if err != nil {
				results = append(results, itemResult{Filename: safeName, Category: chosenCategory, Status: "failed", Error: err.Error()})
				_ = part.Close()
				continue
			}
			if err := part.Close(); err != nil {
				results = append(results, itemResult{Filename: safeName, Category: chosenCategory, Status: "failed", Error: "关闭上传文件失败"})
				continue
			}

			successCount++
			results = append(results, itemResult{
				Filename:        safeName,
				RenamedFilename: archiveResult.RenamedFilename,
				Category:        archiveResult.Category,
				CategoryAbbr:    archiveResult.CategoryAbbr,
				DestinationPath: archiveResult.DestinationPath,
				Status:          "success",
			})
		default:
			_, _ = io.Copy(io.Discard, part)
		}

		_ = part.Close()
	}

	if totalFiles == 0 {
		writeJSONError(w, http.StatusBadRequest, "未选择文件")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"message":      fmt.Sprintf("已处理 %d 个文件，成功 %d 个", totalFiles, successCount),
		"total":        totalFiles,
		"successCount": successCount,
		"results":      results,
	})
}

func (s *Server) handleAITest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var ai archiver.AIConfig
	if err := json.NewDecoder(r.Body).Decode(&ai); err != nil {
		writeJSONError(w, http.StatusBadRequest, "请求格式无效")
		return
	}
	if !aiConfigured(ai) {
		writeJSONError(w, http.StatusBadRequest, "AI 配置不完整，请填写 url/apiKey/modelName")
		return
	}

	content, err := callOpenAICompatible(r.Context(), ai, "请只回复 OK")
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"message": "AI 连通成功",
		"reply":   strings.TrimSpace(content),
	})
}

func (s *Server) handleResolvePath(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rawPath := strings.TrimSpace(r.URL.Query().Get("path"))
	if rawPath == "" {
		writeJSON(w, http.StatusOK, map[string]string{"absolutePath": ""})
		return
	}

	absPath, err := filepath.Abs(rawPath)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "路径解析失败")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"absolutePath": absPath})
}

func aiConfigured(ai archiver.AIConfig) bool {
	return strings.TrimSpace(ai.URL) != "" &&
		strings.TrimSpace(ai.APIKey) != "" &&
		strings.TrimSpace(ai.ModelName) != ""
}

func classifyCategoryByFilename(ctx context.Context, filename string, cfg *archiver.Config) (string, error) {
	if cfg == nil {
		return "", fmt.Errorf("配置为空")
	}
	if len(cfg.Categories) == 0 {
		return "", fmt.Errorf("分类映射为空")
	}

	categories := make([]string, 0, len(cfg.Categories))
	for name := range cfg.Categories {
		categories = append(categories, name)
	}
	sort.Strings(categories)

	respCategory, err := callOpenAICompatible(ctx, cfg.AI, buildClassifyPrompt(filename, categories))
	if err != nil {
		return "", err
	}
	respCategory = strings.TrimSpace(respCategory)
	if respCategory == "" {
		return "", fmt.Errorf("AI 未返回分类结果")
	}
	if _, ok := cfg.Categories[respCategory]; !ok {
		return "", fmt.Errorf("AI 返回了未配置分类: %s", respCategory)
	}
	return respCategory, nil
}

func buildClassifyPrompt(filename string, categories []string) string {
	return fmt.Sprintf(
		"你是文件分类助手。根据文件名，从候选分类中选一个最合适的分类，只输出分类名本身，不要任何解释。\n文件名: %s\n候选分类: %s\n输出要求: 仅输出一个分类名称，必须和候选项完全一致。",
		filename,
		strings.Join(categories, "、"),
	)
}

type chatCompletionsRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionsResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func callOpenAICompatible(ctx context.Context, ai archiver.AIConfig, prompt string) (string, error) {
	url := strings.TrimSpace(ai.URL)
	if !strings.HasSuffix(url, "/chat/completions") {
		url = strings.TrimRight(url, "/") + "/chat/completions"
	}

	reqBody := chatCompletionsRequest{
		Model: ai.ModelName,
		Messages: []chatMessage{
			{Role: "system", Content: "你是一个严谨的文件分类器。"},
			{Role: "user", Content: prompt},
		},
		Temperature: 0,
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("构建 AI 请求失败: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("创建 AI 请求失败: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+ai.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("调用 AI 服务失败: %w", err)
	}
	defer resp.Body.Close()

	var parsed chatCompletionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", fmt.Errorf("解析 AI 响应失败")
	}
	if resp.StatusCode >= 300 {
		if parsed.Error != nil && parsed.Error.Message != "" {
			return "", fmt.Errorf("AI 服务返回错误: %s", parsed.Error.Message)
		}
		return "", fmt.Errorf("AI 服务返回错误: HTTP %d", resp.StatusCode)
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("AI 响应为空")
	}
	return parsed.Choices[0].Message.Content, nil
}

type gzipResponseWriter struct {
	http.ResponseWriter
	writer io.Writer
}

func (w *gzipResponseWriter) Write(p []byte) (int, error) {
	return w.writer.Write(p)
}

func gzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Accept-Encoding")
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		next.ServeHTTP(&gzipResponseWriter{ResponseWriter: w, writer: gz}, r)
	})
}

func buildETag(content []byte) string {
	sum := sha256.Sum256(content)
	return fmt.Sprintf("\"%x\"", sum[:8])
}

func serveEmbeddedHTML(w http.ResponseWriter, r *http.Request, content []byte, etag string) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if strings.TrimSpace(r.Header.Get("If-None-Match")) == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	headers := w.Header()
	headers.Set("Content-Type", "text/html; charset=utf-8")
	headers.Set("Cache-Control", "public, max-age=300")
	headers.Set("ETag", etag)
	_, _ = w.Write(content)
}

func readSmallFormValue(r io.Reader, max int64) string {
	if max <= 0 {
		max = 1024
	}
	b, _ := io.ReadAll(io.LimitReader(r, max))
	return string(b)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
