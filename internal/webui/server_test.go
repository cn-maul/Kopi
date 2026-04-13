package webui

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeJSONBodyRejectsUnknownFields(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest("POST", "/api/config", strings.NewReader(`{"unknown":"value"}`))
	rec := httptest.NewRecorder()

	var payload struct{}
	err := decodeJSONBody(rec, req, &payload)
	if err == nil {
		t.Fatalf("expected decodeJSONBody to reject unknown fields")
	}
}

func TestDecodeJSONBodyRejectsOversizedRequest(t *testing.T) {
	t.Parallel()

	oversized := strings.Repeat("a", maxJSONBodyBytes+16)
	req := httptest.NewRequest("POST", "/api/config", strings.NewReader(`{"x":"`+oversized+`"}`))
	rec := httptest.NewRecorder()

	var payload map[string]string
	err := decodeJSONBody(rec, req, &payload)
	if err == nil {
		t.Fatalf("expected decodeJSONBody to reject oversized payload")
	}
}
