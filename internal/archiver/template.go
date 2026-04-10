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

func renderTemplate(tmpl string, values TemplateValues) (string, error) {
	replacer := strings.NewReplacer(
		"{category_abbr}", values.CategoryAbbr,
		"{yyyymmdd}", values.Date,
		"{filename}", values.Filename,
	)

	rendered := replacer.Replace(tmpl)
	if containsUnknownPlaceholder(rendered) {
		return "", fmt.Errorf("模板包含未识别占位符: %s", tmpl)
	}

	return rendered, nil
}

func containsUnknownPlaceholder(value string) bool {
	return regexp.MustCompile(`\{[^{}]+\}`).MatchString(value)
}
