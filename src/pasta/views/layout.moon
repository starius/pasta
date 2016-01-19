html = require "lapis.html"

class extends html.Widget
  content: =>
    html_5 ->
      head ->
        meta charset: 'utf-8'
        title "Pasta"
      body ->
        if @token
          h1 "Pasta " .. @token
        else
          h1 "Pasta"
        @content_for "inner"
        br!
        a href: "http://github.com/starius/pasta", ->
          text "The source"
        text " of the site is under "
        a href: "https://github.com/starius/pasta/blob/master/LICENSE", ->
          text "the MIT license"
