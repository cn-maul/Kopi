package webui

import (
	"fmt"
	"net/http"
	"strings"
)

func Serve(addr, configPath string) error {
	server := &Server{ConfigPath: configPath}
	return http.ListenAndServe(addr, server.Handler())
}

func buildWebURL(addr string) string {
	trimmed := strings.TrimSpace(addr)
	if trimmed == "" {
		trimmed = ":8080"
	}
	if strings.HasPrefix(trimmed, ":") {
		return "http://localhost" + trimmed
	}
	return fmt.Sprintf("http://%s", trimmed)
}
