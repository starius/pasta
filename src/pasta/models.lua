--[[
Fields of model Pasta:
  * varchar id
  * varchar filename
  * string content
]]

local models = {}

local Model = require("lapis.db.model").Model
local schema = require("lapis.db.schema")

models.Pasta = Model:extend("Pasta")

models.create_schema = function()
    schema.create_table("pasta", {
        {"id", schema.types.varchar},
        {"filename", schema.types.varchar},
        {"content", schema.types.text},
    })
    schema.create_index("pasta", "id")
end

return models
