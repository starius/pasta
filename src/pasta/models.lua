--[[
Pasta's token is a part of URL:
https://pasta.cf/<token>

See also src/migrations.lua

Fields of model Pasta:
  * varchar hash (sha256 of
    config.hash_secret1 .. token .. config.hash_secret2)
  * boolean self_burning
  * varchar filename (not for self-burning pastas)
  * varchar salt (only for self-burning pastas)
  * string content
  * varchar password (used to delete or update; empty if not used or
    sha256(config.password_secret1 .. token .. config.password_secret2))

Filename and content of self-burning pastas are encrypted with
lapis.util.encode_with_secret({content=content, filename=filename}, key),
where key="$salt|$token".
]]

local models = {}

local Model = require("lapis.db.model").Model

models.Pasta = Model:extend("pasta", {
    primary_key = "hash",
})

return models
