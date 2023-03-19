package ui

import (
	_ "embed"
)

//go:embed teams.json
var TeamsJson []byte

//go:embed report.html.tpl
var ReportTemplate string
