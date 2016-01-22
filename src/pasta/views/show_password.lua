local Widget
Widget = require("lapis.html").Widget
local ShowPassword
do
  local _class_0
  local _parent_0 = Widget
  local _base_0 = {
    content = function(self)
      p(function()
        text("Password: ")
        return b(self.password_plain)
      end)
      p(function()
        return font({
          color = "red"
        }, function()
          return text("Write it down, because it will not be shown again!")
        end)
      end)
      text("Your pasta: ")
      return a({
        href = self.pasta_url
      }, function()
        return text(self.pasta_url)
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
    __name = "ShowPassword",
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
  ShowPassword = _class_0
  return _class_0
end
