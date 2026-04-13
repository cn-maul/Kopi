//go:build !tray

package webui

import "fmt"

func RunWithTray(addr, configPath string) error {
	return fmt.Errorf("当前构建未启用托盘支持，请使用 go build -tags tray")
}
