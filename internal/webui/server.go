package webui

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"filearchiver/internal/archiver"
)

type Server struct {
	ConfigPath string
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/settings", s.handleSettings)
	mux.HandleFunc("/api/config", s.handleConfig)
	mux.HandleFunc("/api/archive", s.handleArchive)
	mux.HandleFunc("/api/ai/test", s.handleAITest)
	mux.HandleFunc("/api/path/resolve", s.handleResolvePath)
	return mux
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(indexHTML)
}

func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(settingsHTML)
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

	if err := r.ParseMultipartForm(64 << 20); err != nil {
		writeJSONError(w, http.StatusBadRequest, "上传请求无效")
		return
	}

	category := strings.TrimSpace(r.FormValue("category"))
	useAI := strings.EqualFold(strings.TrimSpace(r.FormValue("useAI")), "true")
	template := strings.TrimSpace(r.FormValue("template"))

	files := r.MultipartForm.File["file"]
	if len(files) == 0 {
		writeJSONError(w, http.StatusBadRequest, "未选择文件")
		return
	}

	workDir, err := os.MkdirTemp("", "archiver-upload-*")
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "创建临时目录失败")
		return
	}
	defer os.RemoveAll(workDir)

	cfg, err := archiver.LoadConfig(s.ConfigPath)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if template == "" {
		template = cfg.TemplatePrefix
	}

	if useAI && !aiConfigured(cfg.AI) {
		writeJSONError(w, http.StatusBadRequest, "AI 配置不完整，请先配置 url/apiKey/modelName")
		return
	}
	if !useAI && category == "" {
		writeJSONError(w, http.StatusBadRequest, "分类不能为空")
		return
	}

	type itemResult struct {
		Filename string `json:"filename"`
		Category string `json:"category,omitempty"`
		Status   string `json:"status"`
		Error    string `json:"error,omitempty"`
	}
	results := make([]itemResult, 0, len(files))
	successCount := 0

	for _, header := range files {
		file, err := header.Open()
		if err != nil {
			results = append(results, itemResult{
				Filename: header.Filename,
				Status:   "failed",
				Error:    "打开上传文件失败",
			})
			continue
		}

		safeName := filepath.Base(header.Filename)
		if safeName == "." || safeName == string(filepath.Separator) || safeName == "" {
			safeName = "uploaded-file"
		}
		localPath := filepath.Join(workDir, safeName)

		target, err := os.Create(localPath)
		if err != nil {
			_ = file.Close()
			results = append(results, itemResult{
				Filename: safeName,
				Status:   "failed",
				Error:    "创建临时文件失败",
			})
			continue
		}

		_, copyErr := io.Copy(target, file)
		closeTargetErr := target.Close()
		closeFileErr := file.Close()
		if copyErr != nil || closeTargetErr != nil || closeFileErr != nil {
			results = append(results, itemResult{
				Filename: safeName,
				Status:   "failed",
				Error:    "写入临时文件失败",
			})
			continue
		}

		chosenCategory := category
		if useAI {
			chosenCategory, err = classifyCategoryByFilename(r.Context(), safeName, cfg)
			if err != nil {
				results = append(results, itemResult{
					Filename: safeName,
					Status:   "failed",
					Error:    err.Error(),
				})
				continue
			}
		}

		if err := archiver.Run(localPath, chosenCategory, template, s.ConfigPath); err != nil {
			results = append(results, itemResult{
				Filename: safeName,
				Category: chosenCategory,
				Status:   "failed",
				Error:    err.Error(),
			})
			continue
		}

		successCount++
		results = append(results, itemResult{
			Filename: safeName,
			Category: chosenCategory,
			Status:   "success",
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"message":      fmt.Sprintf("已处理 %d 个文件，成功 %d 个", len(files), successCount),
		"total":        len(files),
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

	prompt := "请只回复 OK"
	content, err := callOpenAICompatible(r.Context(), ai, prompt)
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

	prompt := buildClassifyPrompt(filename, categories)
	respCategory, err := callOpenAICompatible(ctx, cfg.AI, prompt)
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
		"你是文件分类助手。根据文件名，从候选分类中选一个最合适的分类，只输出分类名称本身，不要任何解释。\n文件名: %s\n候选分类: %s\n输出要求: 仅输出一个分类名称，必须和候选项完全一致。",
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

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取 AI 响应失败: %w", err)
	}

	var parsed chatCompletionsResponse
	if err := json.Unmarshal(respBytes, &parsed); err != nil {
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

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
