local Widget
Widget = require("lapis.html").Widget
local filesize = require("filesize")
local config = require("lapis.config").get()
local blobs = require("pasta.blobs")
local ViewPasta
do
  local _class_0
  local _parent_0 = Widget
  local _base_0 = {
    content = function(self)
      if #self.p_filename > 0 then
        text("File " .. tostring(self.p_filename) .. " / ")
      end
      text("Size " .. tostring(filesize(#self.p_content)) .. " / ")
      a({
        href = self:url_for('index')
      }, function()
        return text('new')
      end)
      if self.p.password == '' then
        text(' / ')
        a({
          href = "https://git.io/vgyko"
        }, function()
          return text("uploader")
        end)
      end
      if not self.p.self_burning then
        local url_params = {
          token = self.token,
          filename = self.p_filename
        }
        text(' / ')
        a({
          href = self:url_for('raw_pasta', url_params)
        }, function()
          return text('raw')
        end)
        text(' / ')
        a({
          href = self:url_for('download_pasta', url_params)
        }, function()
          return text('download')
        end)
      end
      if self.p.password ~= '' then
        text(' / ')
        a({
          href = self:url_for('edit', {
            token = self.token
          })
        }, function()
          return text('edit')
        end)
        text(' / ')
        a({
          href = self:url_for('remove', {
            token = self.token
          })
        }, function()
          return text('remove')
        end)
      end
      br()
      br()
      pre(function()
        return code({
          class = self.ext
        }, self.p_content)
      end)
      return script(function()
        return raw(blobs.apply_highlightjs)
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
