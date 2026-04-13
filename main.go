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
	trayMode := flag.Bool("tray", false, "在托盘模式下启动 Web 页面（需 tray 构建标签）")
	addr := flag.String("addr", ":8082", "Web 服务监听地址")
	flag.Parse()

	if len(os.Args) == 1 {
		*webMode = true
		*trayMode = true
	}

	if *webMode {
		fmt.Printf("Web UI: http://localhost%s\n", *addr)
		var err error
		if *trayMode {
			err = webui.RunWithTray(*addr, *configPath)
		} else {
			err = webui.Serve(*addr, *configPath)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
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
