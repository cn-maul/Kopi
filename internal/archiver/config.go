package archiver

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ArchiveBaseDir string            `yaml:"archiveBaseDir" json:"archiveBaseDir"`
	Categories     map[string]string `yaml:"categories" json:"categories"`
	TemplatePrefix string            `yaml:"templatePrefix" json:"templatePrefix"`
	AI             AIConfig          `yaml:"ai" json:"ai"`
}

type AIConfig struct {
	URL       string `yaml:"url" json:"url"`
	APIKey    string `yaml:"apiKey" json:"apiKey"`
	ModelName string `yaml:"modelName" json:"modelName"`
}

var defaultConfig = Config{
	ArchiveBaseDir: "archive",
	TemplatePrefix: "{category_abbr}-{yyyymmdd}-{filename}",
	Categories: map[string]string{
		"教学": "EDU",
		"财务": "FIN",
		"开发": "DEV",
	},
}

var categoryAbbrPattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

func cloneCategories(src map[string]string) map[string]string {
	if len(src) == 0 {
		return map[string]string{}
	}
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func cloneDefaultConfig() *Config {
	cfg := defaultConfig
	cfg.Categories = cloneCategories(defaultConfig.Categories)
	return &cfg
}

func normalizeConfigDefaults(cfg *Config) {
	if cfg.ArchiveBaseDir == "" {
		cfg.ArchiveBaseDir = defaultConfig.ArchiveBaseDir
	}
	if len(cfg.Categories) == 0 {
		cfg.Categories = cloneCategories(defaultConfig.Categories)
	}
	if strings.TrimSpace(cfg.TemplatePrefix) == "" {
		cfg.TemplatePrefix = defaultConfig.TemplatePrefix
	}
}

func loadConfig(configPath string) (*Config, error) {
	if configPath == "" {
		configPath = "config.yaml"
	}

	cleanPath := filepath.Clean(configPath)
	data, err := os.ReadFile(cleanPath)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := cloneDefaultConfig()
			if err := writeConfigFile(cleanPath, cfg); err != nil {
				return nil, fmt.Errorf("创建默认配置文件失败: %w", err)
			}
			return cfg, nil
		}
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}
	normalizeConfigDefaults(&cfg)
	if err := validateConfig(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func LoadConfig(configPath string) (*Config, error) {
	return loadConfig(configPath)
}

func SaveConfig(configPath string, cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("配置不能为空")
	}

	if configPath == "" {
		configPath = "config.yaml"
	}

	cleanPath := filepath.Clean(configPath)
	normalizeConfigDefaults(cfg)

	aiURLSet := strings.TrimSpace(cfg.AI.URL) != ""
	aiKeySet := strings.TrimSpace(cfg.AI.APIKey) != ""
	aiModelSet := strings.TrimSpace(cfg.AI.ModelName) != ""
	if aiURLSet || aiKeySet || aiModelSet {
		if !(aiURLSet && aiKeySet && aiModelSet) {
			return fmt.Errorf("AI 配置不完整，请同时填写 url、apiKey、modelName")
		}
	}
	if err := validateConfig(cfg); err != nil {
		return err
	}

	return writeConfigFile(cleanPath, cfg)
}

func validateConfig(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("配置不能为空")
	}
	if strings.TrimSpace(cfg.ArchiveBaseDir) == "" {
		return fmt.Errorf("archiveBaseDir 不能为空")
	}
	if len(cfg.Categories) == 0 {
		return fmt.Errorf("分类映射不能为空")
	}
	for name, abbr := range cfg.Categories {
		if strings.TrimSpace(name) == "" {
			return fmt.Errorf("分类名称不能为空")
		}
		if strings.TrimSpace(abbr) == "" {
			return fmt.Errorf("分类缩写不能为空: %s", name)
		}
		if !categoryAbbrPattern.MatchString(abbr) {
			return fmt.Errorf("分类缩写仅允许字母、数字、下划线和中划线: %s", name)
		}
	}
	return nil
}

func writeConfigFile(path string, cfg *Config) error {
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("创建配置目录失败: %w", err)
		}
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("生成配置失败: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}
	return nil
}
