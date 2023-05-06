{{ define "pastGame" }}
{{ if .PostponeReason }}
<p>The game was postponed due to {{ .PostponeReason }} at {{ .Venue.Name }}</p>
{{ else }}
<p>
The {{ .W.Team.Name }} ({{ .W.LeagueRecord.Wins }} - {{ .W.LeagueRecord.Losses }})
beat the {{ .L.Team.Name }} ({{ .L.LeagueRecord.Wins }} - {{ .L.LeagueRecord.Losses }})
{{ .W.Score }} to {{ .L.Score}} {{- if .IsWinnerHome }} at home. {{ else }} on the road. {{ end }}
</p>
{{ if .HasLinescore }}

<table>
	<tr>
	<td></td>
	{{ range $i, $_ := .Linescore.Away.Innings }}
	<td>{{ inc $i }}</td>
	{{ end }}

	<td>R</td>
	<td>H</td>
	<td>E</td>
	</tr>

	{{ template "linescoreTeam" .Linescore.Away }}
	{{ template "linescoreTeam" .Linescore.Home }}
</table>

{{ end }}
{{ if and .HasLinescore .CondensedGameUrl }}<br>{{ end }}
{{ if .CondensedGameUrl }}
<video controls width="650">
	<source src="{{ .CondensedGameUrl }}">
</video>
{{ end }}
{{ end }}
{{ end }}
