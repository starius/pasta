package server

import "html/template"

const style = `
<style>
body {
  font: 15px/1.3 Arial, sans-serif;
}

table {
  margin-top: 0.3em;
  border-collapse: collapse;
}

th,td {
  font-weight: normal;
  padding: 0.5em;
}

h1, h2 {
  margin-top: 5px;
  margin-bottom: 5px;
}
</style>
`

var mainTemplate = template.Must(template.New("main").Parse(`<!DOCTYPE HTML><html lang="en">
<head><meta charset="utf-8"/><meta name="viewport" content="width=device-width, initial-scale=1"/>
<title>Pasta</title>
` + style + `
<link type="image/x-icon" rel="icon" href="/favicon.ico"/>
</head>
<body>
<h1>Pasta</h1>

<table>
<tr>
<td>
  {{ if not .TextTab }}<a href="?tab=text">{{end}}
  Upload text
  {{ if not .TextTab }}</a>{{end}}
</td>
{{ if .AllowFiles }}
<td>
  {{ if not .FileTab }}<a href="?tab=file">{{end}}
  Upload file
  {{ if not .FileTab }}</a>{{end}}
</td>
{{end}}
<td>
  {{ if not .ShortnerTab }}<a href="?tab=shortner">{{end}}
  URL shortner
  {{ if not .ShortnerTab }}</a>{{end}}
</td>
</tr>
</table>

<form action="/api/create" method="POST" enctype="multipart/form-data" autocomplete="off">
<input type="hidden" name="browser" value="on"/>

{{ if .TextTab }}
<table>
    <tr>
        <td>
            <input size="20" name="filename"/>
        </td>
        <td>
            <p>File name (optional)</p>
        </td>
    </tr>
</table>
<textarea cols="80" name="content" rows="24"></textarea>
{{ end }}

{{ if .FileTab }}
<br/>
Select a file:
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;
<input type="file" name="file" id="file"/>
<br/><br/>
Max size: {{.MaxSize}}
<br/><br/>
The name of the file will be available from the link.
{{ end }}

{{ if .ShortnerTab }}
<br/>
Enter URL:
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;
<input type="url" name="content" id="content" size="70"/>
<input type="hidden" name="redirect" id="redirect" value="on"/>
{{ end }}

<br/><br/>
<input type="checkbox" name="self_burning" id="self_burning" {{ if .ForcedBurn }}checked onclick="return false;"{{ end }} />
<label for="self_burning" title="If you select this checkbox, the link will be destroyed after the first access.">Self-burning</label>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;
<input type="checkbox" name="long_id" id="long_id"/>
<label for="long_id" title="If you select this checkbox, the link will be long and very hard to guess or remember.">Secure ID</label>

<br/><br/>
<input type="submit" value="{{ if .ShortnerTab }}Shorten{{ else }}Upload{{ end }}"/>
<br/>
</form>

<br/>

{{.Uploads}} {{ if eq .Uploads "1" }} upload has {{ else }} uploads have {{ end }} been made.

{{ if .Domains }}
  <br/><br/>
  The site runs on the following domains:
  {{ range .Domains }}
    <a href="http://{{ print . }}">{{ print . }}</a>
  {{ end }}
{{ end }}

<br/><br/>
Get the <a href="https://github.com/starius/pasta/tree/master/gopasta">source</a> of the site and install on your own server in 3 seconds: <code>go get github.com/starius/pasta/gopasta</code> .

</body></html>`))

var uploadTemplate = template.Must(template.New("upload").Parse(`<!DOCTYPE HTML><html lang="en">
<head><meta charset="utf-8"/><meta name="viewport" content="width=device-width, initial-scale=1"/>
<title>Pasta</title>
` + style + `
<link type="image/x-icon" rel="icon" href="/favicon.ico"/>
</head>
<body>
<h1>Pasta</h1>

Your {{ if .SelfBurning }} one-time {{ end }} link: <a href="{{ .URL }}">{{ .HumanURL }}</a>

</body></html>`))
