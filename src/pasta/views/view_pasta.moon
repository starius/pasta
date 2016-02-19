import Widget from require "lapis.html"
filesize = require "filesize"

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
    pre @p_content
