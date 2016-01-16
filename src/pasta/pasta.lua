local lapis = require("lapis")
local app = lapis.Application()

app:get("/", function()
    return {render = require "pasta.views.index"}
end)

app:post("create", "/create", function()
    return "Ok"
end)

return app
