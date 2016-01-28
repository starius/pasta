local Widget
Widget = require("lapis.html").Widget
local filesize = require("filesize")
local config = require("lapis.config").get()
local Index
do
  local _class_0
  local _parent_0 = Widget
  local _base_0 = {
    content = function(self)
      return form({
        method = "POST",
        action = self:url_for("create")
      }, function()
        element("table", function()
          return tr(function()
            td(function()
              return input({
                name = "filename",
                size = 20
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
        })
        br()
        text("Max size: " .. tostring(filesize(config.max_pasta_size)))
        br()
        br()
        input({
          type = "radio",
          name = "pasta_type",
          value = "standard",
          id = "pasta_type_standard",
          checked = true
        })
        label({
          ["for"] = 'pasta_type_standard'
        }, 'Standard pasta')
        raw(('&nbsp;'):rep(10))
        input({
          type = "radio",
          name = "pasta_type",
          id = "pasta_type_editable",
          value = "editable"
        })
        label({
          ["for"] = 'pasta_type_editable'
        }, 'Editable pasta')
        raw(('&nbsp;'):rep(10))
        input({
          type = "radio",
          name = "pasta_type",
          id = "pasta_type_self_burning",
          value = "self_burning"
        })
        label({
          ["for"] = 'pasta_type_self_burning'
        }, 'Self-burning')
        br()
        br()
        return input({
          type = "submit",
          value = "Upload"
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
    __name = "Index",
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
  Index = _class_0
  return _class_0
end
