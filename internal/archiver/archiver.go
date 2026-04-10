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

func Run(filePath, category, tmpl, configPath string) error {
	cfg, err := loadConfig(configPath)
	if err != nil {
		return err
	}
	return RunWithConfig(filePath, category, tmpl, cfg)
}

func RunWithConfig(filePath, category, tmpl string, cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("配置为空")
	}

	cleanPath := filepath.Clean(filePath)
	fileInfo, err := os.Stat(cleanPath)
	if err != nil {
		return fmt.Errorf("读取源文件失败: %w", err)
	}
	if fileInfo.IsDir() {
		return fmt.Errorf("源路径不是文件: %s", cleanPath)
	}

	categoryAbbr, ok := cfg.Categories[category]
	if !ok {
		return fmt.Errorf("未知的分类: %s", category)
	}

	filenameWithExt := filepath.Base(cleanPath)
	ext := filepath.Ext(filenameWithExt)
	originalBaseName := strings.TrimSuffix(filenameWithExt, ext)
	targetDir := filepath.Join(cfg.ArchiveBaseDir, categoryAbbr)

	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("创建归档目录失败: %w", err)
	}

	renderedPrefix, err := renderTemplate(tmpl, TemplateValues{
		CategoryAbbr: categoryAbbr,
		Date:         time.Now().Format("20060102"),
		Filename:     originalBaseName,
	})
	if err != nil {
		return err
	}

	version, err := getNextVersion(targetDir, renderedPrefix, ext)
	if err != nil {
		return err
	}

	for {
		renderedName := fmt.Sprintf("%s-v%d%s", renderedPrefix, version, ext)
		destinationPath := filepath.Join(targetDir, renderedName)
		err := copyFileExclusive(cleanPath, destinationPath)
		if err == nil {
			return nil
		}
		if errors.Is(err, os.ErrExist) {
			version++
			continue
		}
		return err
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

	destinationFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return fmt.Errorf("创建目标文件失败: %w", err)
	}

	if _, err := io.Copy(destinationFile, sourceFile); err != nil {
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
