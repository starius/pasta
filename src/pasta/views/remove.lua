local Widget
Widget = require("lapis.html").Widget
local Remove
do
  local _class_0
  local _parent_0 = Widget
  local _base_0 = {
    content = function(self)
      return form({
        method = "POST",
        action = self:url_for("remove2", {
          token = self.token
        })
      }, function()
        element("table", function()
          return tr(function()
            td(function()
              return input({
                name = "password",
                size = 20
              })
            end)
            return td(function()
              return p("Password (required)")
            end)
          end)
        end)
        return input({
          type = "submit",
          value = "Remove pasta"
        })
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
    __name = "Remove",
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
  Remove = _class_0
  return _class_0
end
