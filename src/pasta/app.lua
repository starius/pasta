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
    local text = config.hash_secret1 .. token .. config.hash_secret2
    return crypto.digest('SHA256', text)
end

local function findFreeToken(nwords)
    for i = nwords, nwords * 2 do
        local token = makeToken(i)
        local hash = makeHash(token)
        if not model.Pasta:find(hash) then
            return token
        end
    end
end

local function loadPaste(request)
    request.token = request.params.token
    request.hash = makeHash(request.token)
    request.p = model.Pasta:find(request.hash)
    assert(request.p, "No such pasta")
    request.p_content = request.p.content
end

app:get("schema", "/schema", function()
    model.create_schema()
end)

app:get("index", "/", function()
    return {render = require "pasta.views.index"}
end)

app:post("create", "/create", function(request)
    local token = findFreeToken(config.nwords_short)
    if not token then
        return "No free tokens available"
    end
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
    loadPaste(request)
    return {render = require "pasta.views.view_pasta"}
end)

app:get("raw_pasta", "/:token/raw", function(request)
    loadPaste(request)
    request.res.headers["Content-Type"] = "text/plain"
    return {layout = false, render = require "pasta.views.raw_pasta"}
end)

return app
