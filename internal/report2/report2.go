package report2

import "github.com/0queue/mlb-rss/internal/mlb"

type ReportGenerator struct {
	m *mlb.Mlb
}

func NewReportGenerator(m *mlb.Mlb) ReportGenerator {
	return ReportGenerator{
		m: m,
	}
}

type Report2 struct{}

func GenerateReport() Report2 {
	return Report2{}
}
