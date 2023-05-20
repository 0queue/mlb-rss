<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>{{ .Title }}</title>
  <link rel="icon" type="image/png" sizes="32x32" href="/favicon-32x32.png">
</head>
<body>
<h2>{{ .H2 }}</h2>
{{ template "yesterday" .Yesterday }}
{{ template "upcoming" .Upcoming }}
</body>
</html>