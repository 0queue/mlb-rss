{{ define "yesterday" }}
<strong>Yesterday</strong>

{{ if .PastGames}}
{{ if eq (len .PastGames) 1}}
{{ template "pastGame" (index .PastGames 0) }}
{{ else }}
{{ range $i, $pastGame := .PastGames }}
<i>Game {{ $i }}</i>
{{ template "pastGame" $pastGame}}
{{ end }}
{{ end }}
{{ else }}
<p>The {{ .MyTeam.Name }} did not play yesterday</p>
{{ end }}

<p>For more information go to <a href="{{ .BaseballTheater }}">BaseballTheater</a></p>
{{ end }}
