local html = require("lapis.html")
local filesize = require("filesize")
local view = require("pasta.view")
local css = [[body {
  font: 15px/1.3 Arial, sans-serif;
}

table {
  margin-top: 0.3em;
  border-collapse: collapse;
}

th,td {
  font-weight: normal;
  padding: 0.5em;
}

h1, h2 {
  margin-top: 5px;
  margin-bottom: 5px;
}
]]
do
  local _class_0
  local _parent_0 = html.Widget
  local _base_0 = {
    content = function(self)
      return html_5(function()
        head(function()
          meta({
            charset = 'utf-8'
          })
          title("Pasta")
          return style(css)
        end)
        return body(function()
          if self.token then
            h1("Pasta " .. self.token)
          else
            h1("Pasta")
          end
          if not self.no_new_pasta then
            p(function()
              return a({
                href = self:url_for('index')
              }, function()
                return text('create new')
              end)
            end)
          end
          self:content_for("inner")
          br()
          br()
          a({
            href = "http://github.com/starius/pasta"
          }, function()
            return text("The source")
          end)
          text(" of the site is under ")
          a({
            href = "https://github.com/starius/pasta/blob/master/LICENSE"
          }, function()
            return text("the MIT license")
          end)
          local npastas = view.getNumberOfPastas()
          return p("Number of pastas: " .. tostring(filesize(npastas, {
            unix = true,
            base = 10
          })))
        end)
      end)
    end
  }
  _base_0.__index = _base_0
  setmetatable(_base_0, _parent_0.__base)
  _class_0 = setmetatable({
    __init = function(self, ...)
      return _class_0.__parent.__init(self, ...)
    end,
    __base = _base_0,
    __name = nil,
    __parent = _parent_0
  }, {
    __index = function(cls, name)
      local val = rawget(_base_0, name)
      if val == nil then
        local parent = rawget(cls, "__parent")
        if parent then
          return parent[name]
        end
      else
        return val
      end
    end,
    __call = function(cls, ...)
      local _self_0 = setmetatable({}, _base_0)
      cls.__init(_self_0, ...)
      return _self_0
    end
  })
  _base_0.__class = _class_0
  if _parent_0.__inherited then
    _parent_0.__inherited(_parent_0, _class_0)
  end
  return _class_0
end
