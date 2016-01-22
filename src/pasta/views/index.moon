import Widget from require "lapis.html"
filesize = require "filesize"

config = require("lapis.config").get!

class Index extends Widget
  content: =>
    form method: "POST", action: @url_for("create"), ->
      element "table", -> tr ->
        td -> input name: "filename", size: 20
        td -> p "File name (optional)"
      textarea name: "content", cols: 80, rows: 24
      br!
      text "Max size: #{filesize(config.max_pasta_size)} bytes"
      br!
      br!
      input type: "submit", value: "Upload"
