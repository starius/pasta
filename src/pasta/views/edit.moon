import Widget from require "lapis.html"
filesize = require "filesize"

config = require("lapis.config").get!

class Edit extends Widget
  content: =>
    form method: "POST", action: @url_for("edit2", token: @token), ->
      element "table", ->
        tr ->
          td -> input name: "password", type: "password", size: 20
          td -> p "Password (required)"
        tr ->
          td -> input name: "filename", size: 20, value: @p_filename
          td -> p "File name (optional)"
      textarea name: "content", cols: 80, rows: 24, @p_content
      br!
      text "Max size: #{filesize(config.max_pasta_size)} bytes"

      br!
      br!
      input type: "submit", value: "Update"
