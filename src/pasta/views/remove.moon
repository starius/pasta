import Widget from require "lapis.html"

class Remove extends Widget
  content: =>
    form method: "POST", action: @url_for("remove2", token: @token), ->
      element "table", ->
        tr ->
          td -> input name: "password", type: "password", size: 20
          td -> p "Password (required)"
      input type: "submit", value: "Remove pasta"
