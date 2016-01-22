--[[
Pasta's token is a part of URL:
http://pasta.cf/<token>

Fields of model Pasta:
  * varchar hash (sha256 of
    config.hash_secret1 .. token .. config.hash_secret2)
  * boolean self_burning
  * varchar filename
  * string content
  * varchar password (used to delete or update; empty if not used or
    sha256(config.password_secret1 .. token .. config.password_secret2))

Filename and content of self-burning pastas are encrypted
with lapis.encode_with_secret(text, token).
]]

local models = {}

local Model = require("lapis.db.model").Model
local schema = require("lapis.db.schema")

models.Pasta = Model:extend("pasta", {
    primary_key = "hash",
})

models.create_schema = function()
    schema.create_table("pasta", {
        {"hash", schema.types.varchar},
        {"self_burning", schema.types.boolean},
        {"filename", schema.types.varchar},
        {"content", schema.types.text},
        {"password", schema.types.varchar},
    })
    schema.create_index("pasta", "hash")
end

return models
