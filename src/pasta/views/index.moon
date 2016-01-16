import Widget from require "lapis.html"

class Index extends Widget
  content: =>
    h1 "Pasta"
    form method: "POST", action: @url_for("create"), ->
      textarea name: "text", cols: 80, rows: 24
      br!
      input type: "submit", value: "Upload"
