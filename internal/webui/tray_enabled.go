//go:build tray

package webui

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"github.com/energye/systray"
	trayicon "github.com/energye/systray/icon"
)

func RunWithTray(addr, configPath string) error {
	pageURL := buildWebURL(addr)
	server := &http.Server{
		Addr:    addr,
		Handler: (&Server{ConfigPath: configPath}).Handler(),
	}

	serverErrCh := make(chan error, 1)
	go func() {
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrCh <- err
			systray.Quit()
		}
	}()

	systray.Run(func() {
		systray.SetIcon(trayicon.Data)
		systray.SetTitle("Kopi")
		systray.SetTooltip("Kopi 文件归档")
		systray.SetOnDClick(func(menu systray.IMenu) {
			_ = openBrowser(pageURL)
		})

		openItem := systray.AddMenuItem("打开前端页面", "在浏览器打开 Kopi")
		quitItem := systray.AddMenuItem("退出", "退出 Kopi")
		openItem.Click(func() {
			_ = openBrowser(pageURL)
		})
		quitItem.Click(func() {
			systray.Quit()
		})
	}, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
	})

	select {
	case err := <-serverErrCh:
		return fmt.Errorf("托盘模式下 Web 服务异常退出: %w", err)
	default:
		return nil
	}
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}
