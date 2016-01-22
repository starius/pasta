import Widget from require "lapis.html"

class ShowSelfBurning extends Widget
  content: =>

    p ->
      text "Your pasta: "
      a href: @pasta_url, ->
        text @pasta_url

    p ->
      text "Raw: "
      a href: @pasta_url_raw, ->
        text @pasta_url_raw

    p ->
      text "Download: "
      a href: @pasta_url_download, ->
        text @pasta_url_download

    p -> font color: "red", ->
      text "This pasta will be removed after the first access!"
