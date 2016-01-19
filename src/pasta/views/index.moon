import Widget from require "lapis.html"
config = require("lapis.config").get!

class Index extends Widget
  content: =>
    h1 "Pasta"
    form method: "POST", action: @url_for("create"), ->
      textarea name: "content", cols: 80, rows: 24
      br!
      text "Max size: #{config.max_pasta_size} bytes"
      br!
      input type: "submit", value: "Upload"
