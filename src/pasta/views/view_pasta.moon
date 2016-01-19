import Widget, escape from require "lapis.html"

class ViewPasta extends Widget
  content: =>
    if #@p_filename > 0
      p "File #{@p_filename}"
    a href: @url_for('index'), ->
      text 'new'
    text ' / '
    a href: @url_for('raw_pasta', token: @token, filename: @p_filename), ->
      text 'raw'
    text ' / '
    a href: @url_for('download_pasta', token: @token, filename: @p_filename), ->
      text 'download'
    pre ->
      raw escape @p_content
