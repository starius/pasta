import Widget from require "lapis.html"
filesize = require "filesize"
config = require("lapis.config").get!

get_ext = (filename) ->
  return nil unless filename
  ext = filename\match '%.([^.]+)$'
  if ext and #ext < 10
    return ext

class ViewPasta extends Widget
  content: =>
    if #@p_filename > 0
      text "File #{@p_filename} / "
    text "Size #{filesize(#@p_content)} / "
    a href: @url_for('index'), ->
      text 'new'
    if @p.password == ''
      text ' / '
      a href: "https://git.io/vgyko", ->
        text "uploader"
    if not @p.self_burning
      url_params = token: @token, filename: @p_filename
      text ' / '
      a href: @url_for('raw_pasta', url_params), ->
        text 'raw'
      text ' / '
      a href: @url_for('download_pasta', url_params), ->
        text 'download'
    if @p.password != ''
      text ' / '
      a href: @url_for('edit', token: @token), ->
        text 'edit'
      text ' / '
      a href: @url_for('remove', token: @token), ->
        text 'remove'
    br!
    br!
    ext = get_ext(@p_filename)
    pre ->
      code {class: ext}, @p_content
    if ext
      base_path = config.highlight_js_path
      link rel: 'stylesheet', href: base_path .. 'default.min.css'
      script src: base_path .. 'highlight.min.js'
      script 'hljs.initHighlightingOnLoad()'
