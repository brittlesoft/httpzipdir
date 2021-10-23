package main

const dirlistTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta http-equiv="X-UA-Compatible" content="ie=edge">
  <title>{{ .DirName }}</title>
  <style>
    body {
		font-family: monospace;
		padding: 48px;
	}
	header {
		padding: 4px 16px;
		font-size: 24px;
	}
    ul {
			list-style-type: none;
			margin: 0;
			padding: 20px 0 0 0;
			display: flex;
			flex-wrap: wrap;
    }
    li {
			width: 300px;
			padding: 16px;
		}
		li a {
			display: block;
			overflow: hidden;
			white-space: nowrap;
			text-overflow: ellipsis;
			text-decoration: none;
			transition: opacity 0.25s;
		}
		li span {
			color: #707070;
			font-size: 12px;
		}
		li a:hover {
			opacity: 0.50;
		}
		.dir {
			color: #E91E63;
		}
		.file {
			color: #673AB7;
		}
  </style>
</head>
<body>
	<header>
		{{ .DirName }}
	</header>
	<ul>
		{{ range .Files }}
		<li>
		{{ if .Dir }}
			{{ $name := print .Name "/" }}
			<a class="dir" href="{{PathJoin $.DirName $name }}">{{ $name }}</a>
		{{ else }}
			<a class="file" href="{{ $.DirName }}/{{.Name }}">{{ .Name }}</a>
			<span>{{ .Size }}</span>
		{{ end }}
		</li>
		{{ end }}
  </ul>
</body>
</html>
`
