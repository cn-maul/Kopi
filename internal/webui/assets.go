package webui

import _ "embed"

//go:embed static/index.html
var indexHTML []byte

//go:embed static/settings.html
var settingsHTML []byte
