package main

import (
	"flag"
	"fmt"
	"os"

	"filearchiver/internal/archiver"
	"filearchiver/internal/webui"
)

const defaultTemplate = "{category_abbr}-{yyyymmdd}-{filename}"

func main() {
	filePath := flag.String("f", "", "需要归档的文件路径")
	category := flag.String("c", "", "文件分类")
	template := flag.String("t", defaultTemplate, "文件名前缀模板（版本和扩展名会自动追加）")
	configPath := flag.String("config", "", "配置文件路径 (默认: ./config.yaml)")
	webMode := flag.Bool("web", false, "启动 Web 页面")
	addr := flag.String("addr", ":8080", "Web 服务监听地址")
	trayMode := flag.Bool("tray", true, "启动系统托盘（未启用托盘构建时自动降级）")
	flag.Parse()

	if *webMode {
		fmt.Printf("Web 页面地址: http://localhost%s\n", *addr)
		if *trayMode {
			if err := webui.RunWithTray(*addr, *configPath); err == nil {
				return
			} else {
				fmt.Fprintf(os.Stderr, "托盘模式不可用，已降级为普通 Web 启动: %v\n", err)
			}
		}

		if err := webui.Serve(*addr, *configPath); err != nil {
			fmt.Fprintln(os.Stderr, "Web 服务启动失败:", err)
			os.Exit(1)
		}
		return
	}

	if *filePath == "" || *category == "" {
		flag.Usage()
		fmt.Fprintln(os.Stderr, "-f 和 -c 为必填参数")
		os.Exit(1)
	}

	if err := archiver.Run(*filePath, *category, *template, *configPath); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
