import Widget from require "lapis.html"
filesize = require "filesize"

config = require("lapis.config").get!

class Index extends Widget
  content: =>
    form method: "POST", action: @url_for("create"), ->
      element "table", -> tr ->
        td -> input name: "filename", size: 20
        td -> p "File name (optional)"
      local cols, rows
      if ngx.var.agent_type == 'mobile'
          cols = 45
          rows = 20
        else
          cols = 80
          rows = 24
      textarea name: "content", cols: cols, rows: rows
      br!
      text "Max size: #{filesize(config.max_pasta_size)}"

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
      if ngx.var.agent_type == 'mobile'
        br!
      raw '&nbsp;'\rep 10
      input {
        type: "radio"
        name: "pasta_type"
        id: "pasta_type_self_burning"
        value: "self_burning"
      }
      label for: 'pasta_type_self_burning', 'Self-burning'

      br!
      br!
      input type: "submit", value: "Upload"
