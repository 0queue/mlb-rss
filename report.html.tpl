<h1>Yesterday</h1>
{{- if not .Yesterday }}
<p>The {{ .MyTeam }} did not play yesterday</p>
{{ else }}
<p>The {{ .MyTeam }} {{ .Yesterday.Outcome }} {{ .Yesterday.MyTeamScore }} - {{ .Yesterday.OtherTeamScore }}</p>
{{- end }}

<h1>Upcoming</h1>

<p>todo</p>