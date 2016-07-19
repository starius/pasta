local mnemonic = require("mnemonic")
local arc4random = require("arc4random")
local crypto = require("crypto")
local lru = require("lru")
local yaml = require("yaml")
local ngx = require("ngx")
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

local function apiResponse(object)
    return yaml.dump(object), {
        content_type = 'application/yaml',
        layout = false,
    }
end

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

local function deletePasta(p, token)
    p:delete()
    if cache then
        cache:delete(token)
    end
    if number_of_pastas then
        number_of_pastas = number_of_pastas - 1
    end
end

local function loadPasta(request)
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

local function checkFilename(filename)
    if type(filename) ~= 'string' then
        return nil, "Filename must be a string"
    end
    if #filename > config.max_filename then
        return nil, "Filename is too long. Max " .. config.max_filename
    end
    if filename:match('/') then
        return nil, "Filename must not contain /"
    end
    return true
end

local function makePasta(filename, content, pasta_type)
    local filename_ok, message = checkFilename(filename)
    if type(content) ~= 'string' then
        return nil, "Content must be a string"
    end
    if type(pasta_type) ~= 'string' then
        return nil, "pasta_type must be a string"
    end
    if not filename_ok then
        return filename_ok, message
    end
    local password_hash = ''
    local self_burning = false
    local redirect_to = false
    local nwords = config.nwords.short
    local password_plain
    if pasta_type == 'standard' then
    elseif pasta_type == 'editable' then
        password_plain = makePassword()
        password_hash = makePasswordHash(password_plain)
    elseif pasta_type == 'self_burning' then
        self_burning = true
        nwords = config.nwords.long
    elseif pasta_type == 'url_shortener' then
        filename = ''
        redirect_to = true
    else
        return nil, "Unknown pasta type"
    end
    -- try to find a free token
    local p, token
    for i = nwords, nwords * 2 do
        token = makeToken(i)
        local salt = ''
        if pasta_type == 'self_burning' then
            local info = {
                filename = filename,
                content = content,
            }
            filename = ''
            salt = makeSalt()
            local key = makeKey(salt, token)
            content = encode_with_secret(info, key)
        end
        local _, msg = pcall(function()
            p = model.Pasta:create {
                hash = makeHash(token),
                self_burning = self_burning,
                redirect_to = redirect_to,
                filename = filename,
                salt = salt,
                content = content,
                password = assert(password_hash),
            }
        end)
        if p then
            break
        elseif msg:match('invalid message format') then
            return nil, "Failed to create paste"
        end
    end
    if not p then
        return nil, "No free tokens available"
    end
    if number_of_pastas then
        number_of_pastas = number_of_pastas + 1
    end
    if cache then
        cache:set(token, p, #p.content)
    end
    return {
        token = token,
        password_plain = password_plain,
    }
end

local function getExt(filename)
  if not filename then
    return nil
  end
  local ext = filename:match('%.([^.]+)$')
  if ext and #ext < 10 then
    return ext
  end
end

function view.index(request)
    request.no_new_pasta = true
    return {render = true}
end

function view.createPasta(request)
    local pasta, err = makePasta(
        request.params.filename,
        request.params.content,
        request.params.pasta_type
    )
    if not pasta then
        return {status=ngx.HTTP_BAD_REQUEST}, err
    end
    local url = request:url_for("view_pasta", {token=pasta.token})
    request.token = pasta.token
    request.filename = request.params.filename
    request.pasta_url = request:build_url(url)
    if request.params.pasta_type == 'standard' then
        return {redirect_to = url}
    elseif request.params.pasta_type == 'editable' then
        request.password_plain = pasta.password_plain
        return {render = "show_password"}
    elseif request.params.pasta_type == 'self_burning' then
        return {render = "show_self_burning"}
    elseif request.params.pasta_type == 'url_shortener' then
        return {render = "show_url_shortener"}
    end
end

function view.apiCreatePasta(request)
    local pasta, err = makePasta(
        request.params.filename or '',
        request.params.content,
        request.params.pasta_type
    )
    if not pasta then
        return apiResponse({error = err})
    end
    local result = {password = pasta.password_plain, token=pasta.token}
    pasta.filename = request.params.filename
    result.view = request:build_url(request:url_for("view_pasta", pasta))
    result.raw = request:build_url(request:url_for("raw_pasta", pasta))
    return apiResponse(result)
end

function view.viewPasta(request)
    loadPasta(request)
    if not request.p then
        return {status=ngx.HTTP_NOT_FOUND}, "No such pasta"
    end
    if request.p.redirect_to then
        return {redirect_to = request.p.content}
    end
    request.no_new_pasta = true
    request.ext = getExt(request.p_filename)
    if request.ext ~= 'txt' or request.p_filename == 'CMakeLists.txt' then
        request.highlight = true
    end
    return {render = true}
end

function view.rawPasta(request)
    loadPasta(request)
    if not request.p then
        return {status=ngx.HTTP_NOT_FOUND}, "No such pasta"
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
    loadPasta(request)
    if not request.p then
        return {status=ngx.HTTP_NOT_FOUND}, "No such pasta"
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
    loadPasta(request)
    if not request.p then
        return {status=ngx.HTTP_NOT_FOUND}, "No such pasta"
    end
    if not isEditable(request.p) then
        return {status=ngx.HTTP_FORBIDDEN}, "The pasta is not editable"
    end
    request.no_new_pasta = true
    return {render = true}
end

function view.editPasta2(request)
    loadPasta(request)
    if not request.p then
        return {status=ngx.HTTP_NOT_FOUND}, "No such pasta"
    end
    if not isEditable(request.p) then
        return {status=ngx.HTTP_FORBIDDEN}, "The pasta is not editable"
    end
    if makePasswordHash(request.params.password) ~= request.p.password then
        return {status=ngx.HTTP_FORBIDDEN}, "Wrong password"
    end
    local filename_ok, message = checkFilename(request.params.filename)
    if not filename_ok then
        return {status=ngx.HTTP_BAD_REQUEST}, message
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
    loadPasta(request)
    if not request.p then
        return {status=ngx.HTTP_NOT_FOUND}, "No such pasta"
    end
    if not isEditable(request.p) then
        return {status=ngx.HTTP_FORBIDDEN}, "The pasta is not removable"
    end
    request.no_new_pasta = true
    return {render = true}
end

function view.removePasta2(request)
    loadPasta(request)
    if not request.p then
        return {status=ngx.HTTP_NOT_FOUND}, "No such pasta"
    end
    if not isEditable(request.p) then
        return {status=ngx.HTTP_FORBIDDEN}, "The pasta is not removable"
    end
    if makePasswordHash(request.params.password) ~= request.p.password then
        return {status=ngx.HTTP_FORBIDDEN}, "Wrong password"
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
