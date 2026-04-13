package archiver

import (
	"fmt"
	"regexp"
	"strings"
)

type TemplateValues struct {
	CategoryAbbr string
	Date         string
	Filename     string
}

var (
	knownPlaceholderReplacer = strings.NewReplacer(
		"{category_abbr}", "",
		"{yyyymmdd}", "",
		"{filename}", "",
	)
	unknownPlaceholderPattern = regexp.MustCompile(`\{[^{}]+\}`)
)

func renderTemplate(tmpl string, values TemplateValues) (string, error) {
	if containsUnknownPlaceholder(tmpl) {
		return "", fmt.Errorf("模板包含未识别占位符: %s", tmpl)
	}

	replacer := strings.NewReplacer(
		"{category_abbr}", values.CategoryAbbr,
		"{yyyymmdd}", values.Date,
		"{filename}", values.Filename,
	)

	rendered := replacer.Replace(tmpl)
	return rendered, nil
}

func containsUnknownPlaceholder(value string) bool {
	return unknownPlaceholderPattern.MatchString(knownPlaceholderReplacer.Replace(value))
}

func validateRenderedPrefix(rendered string) error {
	trimmed := strings.TrimSpace(rendered)
	if trimmed == "" {
		return fmt.Errorf("模板渲染结果为空")
	}
	if strings.ContainsAny(trimmed, `/\`) {
		return fmt.Errorf("模板渲染结果不能包含路径分隔符")
	}
	if strings.Contains(trimmed, "..") {
		return fmt.Errorf("模板渲染结果不能包含连续点号")
	}
	if strings.ContainsRune(trimmed, 0) {
		return fmt.Errorf("模板渲染结果包含非法字符")
	}
	return nil
}
