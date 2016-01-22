local Widget
Widget = require("lapis.html").Widget
local filesize = require("filesize")
local config = require("lapis.config").get()
local Edit
do
  local _class_0
  local _parent_0 = Widget
  local _base_0 = {
    content = function(self)
      return form({
        method = "POST",
        action = self:url_for("edit2", {
          token = self.token
        })
      }, function()
        element("table", function()
          tr(function()
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
          return tr(function()
            td(function()
              return input({
                name = "filename",
                size = 20,
                value = self.p_filename
              })
            end)
            return td(function()
              return p("File name (optional)")
            end)
          end)
        end)
        textarea({
          name = "content",
          cols = 80,
          rows = 24
        }, self.p_content)
        br()
        text("Max size: " .. tostring(filesize(config.max_pasta_size)) .. " bytes")
        br()
        br()
        return input({
          type = "submit",
          value = "Update"
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
    __name = "Edit",
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
  Edit = _class_0
  return _class_0
end
