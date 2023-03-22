{{ define "pastGame" }}
{{ if .PostponeReason }}
<p>The game was postponed due to {{ .PostponeReason }} at {{ .Venue.Name }}</p>
{{ else }}
<p>
The {{ .W.Team.Name }} ({{ .W.LeagueRecord.Wins }} - {{ .W.LeagueRecord.Losses }})
beat the {{ .L.Team.Name }} ({{ .L.LeagueRecord.Wins }} - {{ .L.LeagueRecord.Losses }})
{{ .W.Score }} to {{ .L.Score}} {{- if .IsWinnerHome }} at home. {{ else }} on the road. {{ end }}
</p>
{{ end }}
{{ end }}
