import Widget from require "lapis.html"
config = require("lapis.config").get!

class Index extends Widget
  content: =>
    form method: "POST", action: @url_for("create"), ->
      element "table", -> tr ->
        td -> input name: "filename", size: 20
        td -> p "File name"
      textarea name: "content", cols: 80, rows: 24
      br!
      text "Max size: #{config.max_pasta_size} bytes"
      br!
      input type: "submit", value: "Upload"
