# FileArchiver

文件归档工具，支持 CLI 与 Web 界面，自动版本号命名、批量上传、可选 AI 分类。

## 主要功能

- CLI 单文件归档
- Web 拖拽/多选批量上传
- 自动版本号：`-v1`, `-v2`, `-v3`...
- 前缀模板配置（版本号和扩展名固定追加在末尾）
- 分类映射可视化编辑（新增/删除）
- AI 自动分类（按文件名逐个分类）
- 自动生成默认配置文件（`config.yaml`）

## 环境要求

- Go 1.20+

## 快速开始

### 1. 编译

```bash
./scripts/build.sh
```

### 2. 启动 Web

```bash
./scripts/start_web.sh
```

默认地址：`http://localhost:8080`

可选：

```bash
./scripts/start_web.sh :9090 ./config.yaml
```

### 3. CLI 运行

```bash
./scripts/run_cli.sh <文件路径> <分类>
```

示例：

```bash
./scripts/run_cli.sh ./report.pdf 开发
```

## CLI 参数

```bash
./archiver -f <文件路径> -c <分类> [-t 前缀模板] [-config 配置文件]
```

参数说明：

- `-f`：源文件路径（必填）
- `-c`：分类中文名（必填，需在配置中存在）
- `-t`：前缀模板（可选）
- `-config`：配置文件路径（可选，默认 `./config.yaml`）
- `-web`：启动 Web
- `-addr`：Web 监听地址，默认 `:8080`

## 命名规则

最终文件名固定为：

```text
<前缀模板渲染结果>-v<版本号><原扩展名>
```

默认前缀模板：

```text
{category_abbr}-{yyyymmdd}-{filename}
```

支持占位符：

- `{category_abbr}`
- `{yyyymmdd}`
- `{filename}`

## Web 使用说明

- `/`：上传页面（支持拖拽/多选批量上传）
- `/settings`：设置页面（程序配置 + AI 配置）

### 批量上传分类规则

- 勾选 AI：每个文件分别由 AI 判断分类
- 不勾选 AI：所有文件使用当前下拉框选中的同一个分类

## 配置文件

默认配置文件：`./config.yaml`

程序找不到配置文件时会自动创建默认配置。

### 配置示例

```yaml
archiveBaseDir: archive
templatePrefix: '{category_abbr}-{yyyymmdd}-{filename}'
categories:
  开发: DEV
  教学: EDU
  财务: FIN
ai:
  url: ''
  apiKey: ''
  modelName: ''
```

字段说明：

- `archiveBaseDir`：归档根目录
- `templatePrefix`：文件名前缀模板
- `categories`：分类映射（中文名 -> 缩写）
- `ai.url` / `ai.apiKey` / `ai.modelName`：AI 配置（OpenAI 兼容）

## 目录结构

```text
.
├── archiver
├── config.yaml
├── scripts/
│   ├── build.sh
│   ├── start_web.sh
│   └── run_cli.sh
├── internal/
│   ├── archiver/
│   └── webui/
└── archive/
```

## 常见问题

- 目录选择按钮无法返回完整本机路径：浏览器安全限制导致，页面会显示服务端解析后的绝对路径预览；必要时可手动填写绝对路径。
- 勾选 AI 后报错：请先在设置页完整填写 `url`、`apiKey`、`modelName`。
