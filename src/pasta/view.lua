local mnemonic = require("mnemonic")
local arc4random = require("arc4random")
local crypto = require("crypto")
local lru = require("lru")
local urldecode = require("lapis.util").unescape
local encode_with_secret = require("lapis.util.encoding").encode_with_secret
local decode_with_secret = require("lapis.util.encoding").decode_with_secret
local model = require("pasta.models")
local config = require("lapis.config").get()

local view = {}

local cache
if config.pastas_cache then
    cache = lru.new(
        config.pastas_cache.nrecords,
        config.pastas_cache.nbytes
    )
end

local number_of_pastas

local function makeMnemonic(nwords)
    local words = mnemonic.randomWords(nwords, arc4random.random)
    return table.concat(words, '-')
end

local function makeToken(nwords)
    return makeMnemonic(nwords)
end

local function makeHash(token)
    local text = config.hash_secret1 .. token .. config.hash_secret2
    return crypto.digest('SHA256', text)
end

local function makePassword()
    return makeMnemonic(config.nwords.password)
end

local function makePasswordHash(password)
    local text = config.password_secret1 .. password .. config.password_secret2
    return crypto.digest('SHA256', text)
end

local function makeSalt()
    return makeMnemonic(config.nwords.salt)
end

local function makeKey(salt, token)
    return salt .. "|" .. token
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

local function deletePasta(p, token)
    p:delete()
    if cache then
        cache:delete(token)
    end
    if number_of_pastas then
        number_of_pastas = number_of_pastas - 1
    end
end

local function loadPaste(request)
    request.token = request.params.token
    request.p = cache and cache:get(request.token)
    if not request.p then
        local hash = makeHash(request.token)
        request.p = model.Pasta:find(hash)
        if request.p and cache then
            cache:set(request.token, request.p, #request.p.content)
        end
    end
    if request.p then
        if request.p.self_burning then
            local key = makeKey(request.p.salt, request.token)
            local info = decode_with_secret(request.p.content, key)
            request.p_content = info.content
            request.p_filename = info.filename
            deletePasta(request.p, request.token)
        else
            request.p_content = request.p.content
            request.p_filename = request.p.filename
        end
    else
        request.token = nil
    end
end

local function isEditable(p)
    return p.password ~= '' and not p.self_burning
end

function view.schema()
    model.create_schema()
end

function view.index(request)
    request.no_new_pasta = true
    return {render = true}
end

function view.createPasta(request)
    if #request.params.filename > config.max_filename then
        return "Filename is too long. Max " .. config.max_filename
    end
    if request.params.filename:match('/') then
        return "Filename must not contain /"
    end
    local password_hash
    local self_burning = false
    local nwords = config.nwords.short
    if request.params.pasta_type == 'standard' then
        password_hash = ''
    elseif request.params.pasta_type == 'editable' then
        request.password_plain = makePassword()
        password_hash = makePasswordHash(request.password_plain)
    elseif request.params.pasta_type == 'self_burning' then
        password_hash = ''
        self_burning = true
        nwords = config.nwords.long
    else
        return "Unknown pasta type"
    end
    local token = findFreeToken(nwords)
    if not token then
        return "No free tokens available"
    end
    local filename = request.params.filename
    local content = request.params.content
    local salt = ''
    if request.params.pasta_type == 'self_burning' then
        local info = {
            filename = filename,
            content = content,
        }
        filename = ''
        salt = makeSalt()
        local key = makeKey(salt, token)
        content = encode_with_secret(info, key)
    end
    local p = model.Pasta:create {
        hash = makeHash(token),
        self_burning = self_burning,
        filename = filename,
        salt = salt,
        content = content,
        password = assert(password_hash),
    }
    if not p then
        return "Failed to create paste"
    end
    if number_of_pastas then
        number_of_pastas = number_of_pastas + 1
    end
    if cache then
        cache:set(token, p, #p.content)
    end
    local url = request:url_for("view_pasta", {token=token})
    if request.params.pasta_type == 'standard' then
        return {redirect_to = url}
    elseif request.params.pasta_type == 'editable' then
        request.no_new_pasta = true
        request.pasta_url = request:build_url(url)
        return {render = "show_password"}
    elseif request.params.pasta_type == 'self_burning' then
        request.no_new_pasta = true
        request.pasta_url = request:build_url(url)
        local raw_url = request:url_for("raw_pasta", {
            token = token,
            filename = request.params.filename,
        })
        request.pasta_url_raw = request:build_url(raw_url)
        local download_url = request:url_for("download_pasta", {
            token = token,
            filename = request.params.filename,
        })
        request.pasta_url_download = request:build_url(download_url)
        return {render = "show_self_burning"}
    end
end

function view.viewPasta(request)
    loadPaste(request)
    if not request.p then
        return "No such pasta"
    end
    request.no_new_pasta = true
    return {render = true}
end

function view.rawPasta(request)
    loadPaste(request)
    if not request.p then
        return "No such pasta"
    end
    if request.p_filename ~= urldecode(request.params.filename or '') then
        return {
            redirect_to = request:url_for("raw_pasta", {
                token = request.params.token,
                filename = request.p_filename,
            }),
        }
    end
    request.res.headers["Content-Type"] = "text/plain; charset=utf-8"
    return {layout = false, render = "raw_pasta"}
end

function view.downloadPasta(request)
    loadPaste(request)
    if not request.p then
        return "No such pasta"
    end
    if request.p_filename ~= urldecode(request.params.filename or '') then
        return {
            redirect_to = request:url_for("download_pasta", {
                token = request.params.token,
                filename = request.p_filename,
            }),
        }
    end
    request.res.headers["Content-Type"] = "application/octet-stream"
    return {layout = false, render = "raw_pasta"}
end

function view.editPasta(request)
    loadPaste(request)
    if not request.p then
        return "No such pasta"
    end
    if not isEditable(request.p) then
        return "The pasta is not editable"
    end
    request.no_new_pasta = true
    return {render = true}
end

function view.editPasta2(request)
    loadPaste(request)
    if not request.p then
        return "No such pasta"
    end
    if not isEditable(request.p) then
        return "The pasta is not editable"
    end
    if makePasswordHash(request.params.password) ~= request.p.password then
        return "Wrong password"
    end
    if #request.params.filename > config.max_filename then
        return "Filename is too long. Max " .. config.max_filename
    end
    request.p:update {
        filename = request.params.filename,
        content = request.params.content,
    }
    if cache then
        cache:set(request.token, request.p, #request.p.content)
    end
    local url = request:url_for("view_pasta", {token=request.token})
    return {redirect_to = url}
end

function view.removePasta(request)
    loadPaste(request)
    if not request.p then
        return "No such pasta"
    end
    if not isEditable(request.p) then
        return "The pasta is not removable"
    end
    request.no_new_pasta = true
    return {render = true}
end

function view.removePasta2(request)
    loadPaste(request)
    if not request.p then
        return "No such pasta"
    end
    if not isEditable(request.p) then
        return "The pasta is not removable"
    end
    if makePasswordHash(request.params.password) ~= request.p.password then
        return "Wrong password"
    end
    deletePasta(request.p, request.token)
    return {redirect_to = request:url_for("index")}
end

function view.getNumberOfPastas()
    if not number_of_pastas then
        number_of_pastas = model.Pasta:count()
    end
    return number_of_pastas
end

return view
