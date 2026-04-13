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
