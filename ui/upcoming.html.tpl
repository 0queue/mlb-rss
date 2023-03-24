{{ define "upcoming" }}
<strong>Upcoming</strong>

<table>
	<tr>
		{{ range .FutureDays }}
		<th>{{ .DayAbbr }}</th>
		{{ end }}
	</tr>

	<tr>
		{{ range .FutureDays}}
		{{ if .Games }}
		<td>
			{{ range $i, $g := .Games }}
			{{ if $i }}<br>{{ end }}
			{{ if not $g.IsMyTeamHome }}@{{ end }}{{ $g.AgainstAbbr }}
			{{ end }}
		</td>
		{{ else }}
		<td>💤</td>
		{{ end }}
		{{ end }}
	</tr>

	<tr>
		{{ range .FutureDays }}
		{{ if .Games }}
		<td>
		{{ range $i, $g := .Games }}
			{{ if $i }}<br>{{ end }}
			{{ $g.GameTimeLocal }}
		{{ end }}
		</td>
		{{ else }}
		<td></td>
		{{ end }}
		{{ end }}
	</tr>
</table>

<!-- yeah yeah this doesn't work in miniflux (unless I trust the site?) -->
<p style="font-size: small; text-align: right;">TZ={{ .Timezone }}</p>

{{ end }}
