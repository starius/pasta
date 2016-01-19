local Widget, escape
do
  local _obj_0 = require("lapis.html")
  Widget, escape = _obj_0.Widget, _obj_0.escape
end
local ViewPasta
do
  local _class_0
  local _parent_0 = Widget
  local _base_0 = {
    content = function(self)
      h1("Pasta " .. self.token)
      p("File " .. tostring(self.p_filename))
      a({
        href = self:url_for('raw_pasta', {
          token = self.token
        })
      }, function()
        return text('raw')
      end)
      text(' / ')
      a({
        href = self:url_for('download_pasta', {
          token = self.token
        })
      }, function()
        return text('download')
      end)
      return pre(function()
        return raw(escape(self.p_content))
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
    __name = "ViewPasta",
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
  ViewPasta = _class_0
  return _class_0
end
