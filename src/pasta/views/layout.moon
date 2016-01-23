html = require "lapis.html"
filesize = require "filesize"

view = require "pasta.view"

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
        if not @no_new_pasta
          p -> a href: @url_for('index'), ->
            text 'create new'
        @content_for "inner"
        br!
        br!
        a href: "http://github.com/starius/pasta", ->
          text "The source"
        text " of the site is under "
        a href: "https://github.com/starius/pasta/blob/master/LICENSE", ->
          text "the MIT license"
        npastas = view.getNumberOfPastas()
        p "Number of pastas: #{filesize(npastas, unix: true, base: 10)}"
