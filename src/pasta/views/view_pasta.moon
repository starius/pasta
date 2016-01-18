import Widget, escape from require "lapis.html"

class ViewPasta extends Widget
  content: =>
    h1 "Pasta " .. @token
    pre ->
      raw escape @p_content
