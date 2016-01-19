local mnemonic = require("mnemonic")
local arc4random = require("arc4random")
local crypto = require("crypto")
local ngx = require("ngx")
local lapis = require("lapis")
local urldecode = require("lapis.util").unescape
local model = require("pasta.models")
local config = require("lapis.config").get()
local app = lapis.Application()

app.layout = require("pasta.views.layout")
app.views_prefix = "pasta.views"

if not config.print_stack_to_browser then
    -- http://leafo.net/lapis/reference/actions.html
    app.handle_error = function(self, _, _)
        ngx.say("There was an error...")
    end
end

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
    local hash = makeHash(request.token)
    request.p = model.Pasta:find(hash)
    assert(request.p, "No such pasta")
    request.p_content = request.p.content
    request.p_filename = request.p.filename
end

app:get("schema", "/schema", function()
    model.create_schema()
end)

app:get("index", "/", function()
    return {render = true}
end)

app:post("create", "/create", function(request)
    if #request.params.filename > config.max_filename then
        return "Filename is too long. Max " .. config.max_filename
    end
    local token = findFreeToken(config.nwords_short)
    if not token then
        return "No free tokens available"
    end
    local p = model.Pasta:create {
        hash = makeHash(token),
        self_burning = false,
        filename = request.params.filename,
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
    return {render = true}
end)

local function rawPasta(request)
    loadPaste(request)
    if request.p.filename ~= urldecode(request.params.filename or '') then
        return {
            redirect_to = request:url_for("raw_pasta", {
                token = request.params.token,
                filename = request.p.filename,
            }),
        }
    end
    request.res.headers["Content-Type"] = "text/plain; charset=utf-8"
    return {layout = false, render = "raw_pasta"}
end

app:get("raw_pasta0", "/:token/raw", rawPasta)
app:get("raw_pasta", "/:token/raw/:filename", rawPasta)

local function downloadPasta(request)
    loadPaste(request)
    if request.p.filename ~= urldecode(request.params.filename or '') then
        return {
            redirect_to = request:url_for("download_pasta", {
                token = request.params.token,
                filename = request.p.filename,
            }),
        }
    end
    request.res.headers["Content-Type"] = "application/octet-stream"
    return {layout = false, render = "raw_pasta"}
end

app:get("download_pasta0", "/:token/download", downloadPasta)
app:get("download_pasta", "/:token/download/:filename", downloadPasta)

return app
