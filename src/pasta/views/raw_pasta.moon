import Widget from require "lapis.html"

class RawPasta extends Widget
  content: =>
    raw @p_content
