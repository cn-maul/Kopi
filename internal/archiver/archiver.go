package archiver

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type VersionCache struct {
	next map[string]int
}

type ArchiveResult struct {
	OriginalFilename string `json:"originalFilename"`
	RenamedFilename  string `json:"renamedFilename"`
	Category         string `json:"category"`
	CategoryAbbr     string `json:"categoryAbbr"`
	DestinationPath  string `json:"destinationPath"`
}

type archiveTargetMeta struct {
	targetDir      string
	renderedPrefix string
	ext            string
	categoryAbbr   string
}

func NewVersionCache() *VersionCache {
	return &VersionCache{next: make(map[string]int)}
}

func Run(filePath, category, tmpl, configPath string) error {
	cfg, err := loadConfig(configPath)
	if err != nil {
		return err
	}
	_, err = RunWithConfigResult(filePath, category, tmpl, cfg)
	return err
}

func RunWithConfig(filePath, category, tmpl string, cfg *Config) error {
	_, err := RunWithConfigResult(filePath, category, tmpl, cfg)
	return err
}

func RunWithConfigResult(filePath, category, tmpl string, cfg *Config) (ArchiveResult, error) {
	if cfg == nil {
		return ArchiveResult{}, fmt.Errorf("配置为空")
	}

	cleanPath := filepath.Clean(filePath)
	fileInfo, err := os.Stat(cleanPath)
	if err != nil {
		return ArchiveResult{}, fmt.Errorf("读取源文件失败: %w", err)
	}
	if fileInfo.IsDir() {
		return ArchiveResult{}, fmt.Errorf("源路径不是文件: %s", cleanPath)
	}

	filenameWithExt := filepath.Base(cleanPath)
	return archiveWithWriter(filenameWithExt, category, tmpl, cfg, nil, func(dst string) error {
		return copyFileExclusive(cleanPath, dst)
	})
}

func RunReaderWithConfig(reader io.Reader, filenameWithExt, category, tmpl string, cfg *Config, versionCache *VersionCache) error {
	_, err := RunReaderWithConfigResult(reader, filenameWithExt, category, tmpl, cfg, versionCache)
	return err
}

func RunReaderWithConfigResult(reader io.Reader, filenameWithExt, category, tmpl string, cfg *Config, versionCache *VersionCache) (ArchiveResult, error) {
	if cfg == nil {
		return ArchiveResult{}, fmt.Errorf("配置为空")
	}
	if reader == nil {
		return ArchiveResult{}, fmt.Errorf("文件内容为空")
	}

	baseName := filepath.Base(strings.TrimSpace(filenameWithExt))
	if baseName == "" || baseName == "." || baseName == string(filepath.Separator) {
		return ArchiveResult{}, fmt.Errorf("文件名不合法")
	}

	return archiveWithWriter(baseName, category, tmpl, cfg, versionCache, func(dst string) error {
		return copyReaderExclusive(reader, dst)
	})
}

func archiveWithWriter(filenameWithExt, category, tmpl string, cfg *Config, versionCache *VersionCache, writer func(dst string) error) (ArchiveResult, error) {
	meta, err := buildTargetMeta(category, tmpl, cfg, filenameWithExt)
	if err != nil {
		return ArchiveResult{}, err
	}

	version, err := getCachedNextVersion(meta.targetDir, meta.renderedPrefix, meta.ext, versionCache)
	if err != nil {
		return ArchiveResult{}, err
	}
	versionKey := buildVersionKey(meta.targetDir, meta.renderedPrefix, meta.ext)

	for {
		renderedName := fmt.Sprintf("%s-v%d%s", meta.renderedPrefix, version, meta.ext)
		destinationPath := filepath.Join(meta.targetDir, renderedName)
		err := writer(destinationPath)
		if err == nil {
			bumpCachedNextVersion(versionCache, versionKey, version+1)
			return ArchiveResult{
				OriginalFilename: filenameWithExt,
				RenamedFilename:  renderedName,
				Category:         category,
				CategoryAbbr:     meta.categoryAbbr,
				DestinationPath:  destinationPath,
			}, nil
		}
		if errors.Is(err, os.ErrExist) {
			version++
			continue
		}
		return ArchiveResult{}, err
	}
}

func buildTargetMeta(category, tmpl string, cfg *Config, filenameWithExt string) (archiveTargetMeta, error) {
	categoryAbbr, ok := cfg.Categories[category]
	if !ok {
		return archiveTargetMeta{}, fmt.Errorf("未知的分类: %s", category)
	}

	ext := filepath.Ext(filenameWithExt)
	originalBaseName := strings.TrimSuffix(filenameWithExt, ext)
	targetDir := filepath.Join(cfg.ArchiveBaseDir, categoryAbbr)

	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return archiveTargetMeta{}, fmt.Errorf("创建归档目录失败: %w", err)
	}

	renderedPrefix, err := renderTemplate(tmpl, TemplateValues{
		CategoryAbbr: categoryAbbr,
		Date:         time.Now().Format("20060102"),
		Filename:     originalBaseName,
	})
	if err != nil {
		return archiveTargetMeta{}, err
	}

	return archiveTargetMeta{
		targetDir:      targetDir,
		renderedPrefix: renderedPrefix,
		ext:            ext,
		categoryAbbr:   categoryAbbr,
	}, nil
}

func getCachedNextVersion(targetDir, renderedPrefix, ext string, versionCache *VersionCache) (int, error) {
	if versionCache == nil {
		return getNextVersion(targetDir, renderedPrefix, ext)
	}

	key := buildVersionKey(targetDir, renderedPrefix, ext)
	if next, ok := versionCache.next[key]; ok {
		versionCache.next[key] = next + 1
		return next, nil
	}

	next, err := getNextVersion(targetDir, renderedPrefix, ext)
	if err != nil {
		return 0, err
	}
	versionCache.next[key] = next + 1
	return next, nil
}

func buildVersionKey(targetDir, renderedPrefix, ext string) string {
	return targetDir + "\n" + renderedPrefix + "\n" + ext
}

func bumpCachedNextVersion(versionCache *VersionCache, key string, next int) {
	if versionCache == nil {
		return
	}
	if current, ok := versionCache.next[key]; !ok || next > current {
		versionCache.next[key] = next
	}
}

func getNextVersion(targetDir, renderedPrefix, ext string) (int, error) {
	entries, err := os.ReadDir(targetDir)
	if err != nil {
		return 0, fmt.Errorf("扫描归档目录失败: %w", err)
	}

	pattern := regexp.MustCompile(`^` + regexp.QuoteMeta(renderedPrefix) + `-v(\d+)` + regexp.QuoteMeta(ext) + `$`)
	maxVersion := 0

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		matches := pattern.FindStringSubmatch(entry.Name())
		if len(matches) != 2 {
			continue
		}

		version, err := strconv.Atoi(matches[1])
		if err != nil {
			continue
		}
		if version > maxVersion {
			maxVersion = version
		}
	}

	return maxVersion + 1, nil
}

func copyFileExclusive(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("打开源文件失败: %w", err)
	}
	defer sourceFile.Close()

	return copyReaderExclusive(sourceFile, dst)
}

func copyReaderExclusive(source io.Reader, dst string) error {
	destinationFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return fmt.Errorf("创建目标文件失败: %w", err)
	}

	if _, err := io.Copy(destinationFile, source); err != nil {
		_ = destinationFile.Close()
		_ = os.Remove(dst)
		return fmt.Errorf("复制文件失败: %w", err)
	}

	if err := destinationFile.Sync(); err != nil {
		_ = destinationFile.Close()
		_ = os.Remove(dst)
		return fmt.Errorf("刷新目标文件失败: %w", err)
	}
	if err := destinationFile.Close(); err != nil {
		_ = os.Remove(dst)
		return fmt.Errorf("关闭目标文件失败: %w", err)
	}

	return nil
}
