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

<td><b>{{ .Runs }}</b></td>
<td>{{ .Hits }}</td>
<td>{{ .Errors }}</td>

</tr>
{{ end }}