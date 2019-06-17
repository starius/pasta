package server

const mainTemplate = `<!DOCTYPE HTML><html lang="en">
<head><meta charset="utf-8"/><meta name="viewport" content="width=device-width, initial-scale=1"/>
<title>Pasta</title>
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
<link type="image/x-icon" rel="icon" href="/favicon.ico"/>
</head>
<body>
<h1>Pasta</h1><form action="/api/create" method="POST" enctype="multipart/form-data">
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
<br/>
Max size: {{.MaxSize}}

<br/><br/>
or select a file: <input type="file" name="file" id="file">

<br/><br/>
<input type="checkbox" name="self_burning" id="self_burning"/>
<label for="self_burning">Self-burning</label>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;
<input type="checkbox" name="long_id" id="long_id"/>
<label for="long_id">Secure ID</label>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;
<input type="checkbox" name="redirect" id="redirect"/>
<label for="redirect">URL shortener</label>

<br/><br/>
<input type="submit" value="Upload"/>
</form>
<br/><br/>
Number of uploads: {{.Uploads}}.
</body></html>`