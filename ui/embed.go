package ui

import (
	"embed"
)

//go:embed *.html.tpl
var ReportTemplates embed.FS

//go:embed favicon-32x32.png
var Favicon []byte
