local lapis = require("lapis")
local model = require("pasta.models")
local app = lapis.Application()

app:get("schema", "/schema", function()
    model.create_schema()
end)

app:get("index", "/", function()
    return {render = require "pasta.views.index"}
end)

app:post("create", "/create", function()
    return "Ok"
end)

return app
