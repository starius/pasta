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
      input {
        type: "radio"
        name: "pasta_type"
        value: "standard"
        id: "pasta_type_standard"
        checked: true
      }
      label for: 'pasta_type_standard', 'Standard pasta'
      raw '&nbsp;'\rep 10
      input {
        type: "radio"
        name: "pasta_type"
        id: "pasta_type_editable"
        value: "editable"
      }
      label for: 'pasta_type_editable', 'Editable pasta'

      br!
      br!
      input type: "submit", value: "Upload"
