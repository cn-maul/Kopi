package archiver

import (
	"strings"
	"testing"
)

func TestRenderTemplateRejectsUnknownPlaceholder(t *testing.T) {
	t.Parallel()

	_, err := renderTemplate("{category_abbr}-{unknown}", TemplateValues{
		CategoryAbbr: "DEV",
		Date:         "20260413",
		Filename:     "demo",
	})
	if err == nil {
		t.Fatalf("expected renderTemplate to reject unknown placeholder")
	}
}

func TestValidateRenderedPrefixRejectsPathSeparators(t *testing.T) {
	t.Parallel()

	err := validateRenderedPrefix("../unsafe")
	if err == nil {
		t.Fatalf("expected validateRenderedPrefix to fail")
	}
	if !strings.Contains(err.Error(), "路径分隔符") && !strings.Contains(err.Error(), "连续点号") {
		t.Fatalf("unexpected error: %v", err)
	}
}
