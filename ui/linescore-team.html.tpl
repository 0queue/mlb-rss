{{ define "linescoreTeam" }}
<tr>
<td>{{ .Abbr }}</td>

{{ range .Innings }}
{{ if lt . 0 }}
<td>x</td>
{{ else }}
<td>{{ . }}</td>
{{ end }}
{{ end }}

<td><strong>{{ .Runs }}</strong></td>
<td>{{ .Hits }}</td>
<td>{{ .Errors }}</td>

</tr>
{{ end }}