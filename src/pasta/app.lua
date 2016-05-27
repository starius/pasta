local ngx = require("ngx")
local lapis = require("lapis")
local config = require("lapis.config").get()

local app = lapis.Application()
app.layout = require("pasta.views.layout")
app.views_prefix = "pasta.views"

local view = require("pasta.view")

if not config.print_stack_to_browser then
    -- http://leafo.net/lapis/reference/actions.html
    app.handle_error = function(_, _, _)
        ngx.status = ngx.HTTP_INTERNAL_SERVER_ERROR
        ngx.say("There was an error...")
    end
end

app:get("index", "/", view.index)

app:post("create", "/pasta/create", view.createPasta)
app:post("api_create", "/api/create", view.apiCreatePasta)

app:get("view_pasta", "/:token", view.viewPasta)

app:get("raw_pasta0", "/:token/raw", view.rawPasta)
app:get("raw_pasta", "/:token/raw/:filename", view.rawPasta)
-- https://stackoverflow.com/a/216777

app:get("download_pasta0", "/:token/download", view.downloadPasta)
app:get("download_pasta", "/:token/download/:filename", view.downloadPasta)

app:get("edit", "/:token/edit", view.editPasta)

app:post("edit2", "/:token/edit2", view.editPasta2)

app:get("remove", "/:token/remove", view.removePasta)

app:post("remove2", "/:token/remove2", view.removePasta2)

return app
