import Widget from require "lapis.html"

class ShowSelfBurning extends Widget
  content: =>

    p ->
      text "Your pasta: "
      a href: @pasta_url, ->
        text @pasta_url

    pasta = {token: @token, filename: @filename}

    pasta_url_raw = @build_url(@url_for("raw_pasta", pasta))
    p ->
      text "Raw: "
      a href: pasta_url_raw, ->
        text pasta_url_raw

    pasta_url_download = @build_url(@url_for("download_pasta", pasta))
    p ->
      text "Download: "
      a href: pasta_url_download, ->
        text pasta_url_download

    p -> font color: "red", ->
      text "This pasta will be removed after the first access!"
