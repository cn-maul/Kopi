package archiver

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfigCreatesDefaultFileWhenMissing(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}
	if cfg == nil {
		t.Fatalf("LoadConfig returned nil config")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("expected config file to be created, got error: %v", err)
	}
	if len(data) == 0 {
		t.Fatalf("created config file is empty")
	}
}

func TestSaveConfigRejectsInvalidCategoryAbbreviation(t *testing.T) {
	t.Parallel()

	cfg := &Config{
		ArchiveBaseDir: "archive",
		TemplatePrefix: "{category_abbr}-{yyyymmdd}-{filename}",
		Categories: map[string]string{
			"开发": "../DEV",
		},
	}

	err := SaveConfig(filepath.Join(t.TempDir(), "config.yaml"), cfg)
	if err == nil {
		t.Fatalf("expected SaveConfig to reject invalid abbreviation")
	}
	if !strings.Contains(err.Error(), "分类缩写") {
		t.Fatalf("unexpected error: %v", err)
	}
}
