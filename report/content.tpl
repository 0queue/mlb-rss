<strong>Yesterday</strong>
{{- if .Postpone }}
<p>The game was postponed due to {{ .Postpone.Reason }} at {{ .Postpone.Where }}</p>
{{- else }}
{{- if .Yesterday }}
{{with .Yesterday }}
<p>The {{ .WinningTeam.Team.Name }} ({{ .WinningTeam.LeagueRecord.Wins }}-{{ .WinningTeam.LeagueRecord.Losses }}) beat the {{ .LosingTeam.Team.Name }} ({{ .LosingTeam.LeagueRecord.Wins }}-{{ .LosingTeam.LeagueRecord.Losses }}) {{ .WinningTeam.Score }} to {{ .LosingTeam.Score }} {{ .Where }}</p>
{{- else }}
<p>The {{ .Team.Name }} did not play yesterday</p>
{{- end }}
{{- end }}
{{- end }}

<p>See yesterday's games at <a href="{{ .BaseballTheater }}">Baseball Theater</a></p>

<strong>Upcoming</strong>

<table>

<tr>
{{- range .UpcomingDayAbbr }}
<th>{{ . }}</th>
{{- end }}
</tr>

<tr>
{{- range .UpcomingInfos }}
{{- if . }}
<td>{{ if not .IsHome }}@ {{ end }}{{ .AgainstAbbr }}</td>
{{- else }}
<td>💤</td>
{{- end }}
{{- end }}
</tr>

<tr>
{{- range .UpcomingTimes }}
{{- if . }}
<td>{{ . }}</td>
{{- else }}
<td></td>
{{- end }}
{{- end }}
</tr>

</table>

<p style="font-size: small; text-align: right;">TZ={{ .UpcomingTimezone }}</p>