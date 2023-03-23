package ui

import (
	"embed"
)

//go:embed *.html.tpl
var ReportTemplates embed.FS
