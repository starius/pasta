local mnemonic = require("mnemonic")
local arc4random = require("arc4random")
local crypto = require("crypto")
local lapis = require("lapis")
local model = require("pasta.models")
local config = require("lapis.config").get()
local app = lapis.Application()

local function makeToken(nwords)
    local words = mnemonic.randomWords(nwords, arc4random.random)
    return table.concat(words, '-')
end

local function makeHash(token)
    return crypto.digest('SHA256', token)
end

app:get("schema", "/schema", function()
    model.create_schema()
end)

app:get("index", "/", function()
    return {render = require "pasta.views.index"}
end)

app:post("create", "/create", function(request)
    local token = makeToken(config.nwords_short)
    local p = model.Pasta:create {
        hash = makeHash(token),
        self_burning = false,
        filename = 'TODO.txt',
        content = request.params.content,
        password = 'TODO',
    }
    if not p then
        return "Failed to create paste"
    end
    local url = request:url_for("view_pasta", {token=token})
    return {redirect_to = url}
end)

app:get("view_pasta", "/:token", function(request)
    request.token = request.params.token
    request.hash = makeHash(request.token)
    request.p = model.Pasta:find(request.hash)
    if not request.p then
        return "No such pasta"
    end
    request.p_content = request.p.content
    return {render = require "pasta.views.view_pasta"}
end)

return app
