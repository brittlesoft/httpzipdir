package main

const dirlistTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>{{ .DirName }}</title>
  <style type="text/css">
  	body { background-color: #F5F5F5; }
  	h2#title { margin-bottom: 12px; }
  	a, a:active { text-decoration: none; color: blue; }
  	a:visited { color: #48468F; }
  	a:hover, a:focus { text-decoration: underline; color: red; }
  	table { margin-left: 12px; }
  	th, td { font: 90% monospace; text-align: left; }
  	th { font-weight: bold; padding-right: 14px; padding-bottom: 3px; }
  	td { padding-right: 14px; }
  	td.size, th#size { text-align: right; }
  	#dirlist { background-color: white; border-top: 1px solid #646464; border-bottom: 1px solid #646464; padding-top: 10px; padding-bottom: 14px; }
  	div#footer { font: 90% monospace; color: #787878; padding-top: 4px; }
  	a.sortheader { color: black; text-decoration: none; display: block; }
  	span.sortarrow { text-decoration: none; }
  </style>
</head>
<body>
	<h2 id="title">Index of {{ .DirName}}</h2>
	<div id="dirlist">
		<table summary="Directory Listing" cellpadding="0" cellspacing="0" class="sort">
			<thead><tr><th id="name">Name:</th><th id="modified" class="int">Last Modified:</th><th id="size" class="int">Size:</th></tr></thead>
			<tbody>
			{{- if ne .DirName "/" }}
			<tr><td><a href="../">Parent Directory</a></td><td class="modified" val="0"></td><td class="size" val="0">-</td></tr>
			{{- end }}
			{{- range .Files }}
			  {{- $modtime := .ModTime | date "2006-01-02 15:04:05 -0700" }}
			  {{- if .Dir }}
			  {{- $name := print .Name "/" }}
			  <tr><td><a href="{{PathJoin $.DirName $name }}/">{{ $name }}</a></td><td class="modified" val="{{ .ModTime | unixEpoch}}">{{$modtime}}</td><td class="size" val="0">-</td></tr>
			  {{- if $.ZipDir }}
			  <tr><td><a href="{{PathJoin $.DirName .Name }}.zip">{{ .Name }}.zip</a></td><td class="modified" val="{{ .ModTime | unixEpoch}}">{{$modtime}}</td><td class="size" val="0">?</td></tr>
			  {{- end }}
			  {{- else }}
			  <tr><td><a href="{{PathJoin $.DirName .Name }}">{{ .Name }}</a></td><td class="modified" val="{{ .ModTime | unixEpoch}}">{{$modtime}}</td><td class="size" val="0">{{.Size}}</td></tr>
			  {{- end }}
			  {{- end}}
			</tbody>
		</table>
	</div>
</body>
</html>
`
