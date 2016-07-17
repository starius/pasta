import Widget from require "lapis.html"

class ShowUrlShortener extends Widget
  content: =>

    p ->
      text "Your pasta: "
      a href: @pasta_url, ->
        text @pasta_url
