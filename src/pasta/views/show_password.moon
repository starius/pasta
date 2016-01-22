import Widget from require "lapis.html"

class ShowPassword extends Widget
  content: =>
    p ->
      text "Password: "
      b @password_plain
    p -> font color: "red", ->
      text "Write it down, because it will not be shown again!"
    text "Your pasta: "
    a href: @pasta_url, ->
      text @pasta_url
