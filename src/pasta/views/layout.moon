html = require "lapis.html"
filesize = require "filesize"

view = require "pasta.view"
import main_css from require "pasta.blobs"

class extends html.Widget
  content: =>
    html_5 ->
      head ->
        meta charset: 'utf-8'
        meta name: 'viewport', content: 'width=device-width, initial-scale=1'
        title "Pasta"
        style main_css
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
        npastas = view.getNumberOfPastas()
        text "Number of pastas: #{filesize(npastas, unix: true, base: 10)}. "
        text "Mirrors: "
        a href: "https://pasta.cf", ->
          text "pasta.cf"
        text ", "
        a href: "http://pastagdsp33j7aoq.onion", ->
          text "pastagdsp33j7aoq.onion"
        text "."
        br!
        a href: "http://github.com/starius/pasta", ->
          text "The source"
        text " of the site is under "
        a href: "https://github.com/starius/pasta/blob/master/LICENSE", ->
          text "the MIT license"
        text "."
