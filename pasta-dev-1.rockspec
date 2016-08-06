package = "pasta"
version = "dev-1"
source = {
    url = "git://github.com/starius/pasta.git"
}
description = {
    summary = "Pastebin-like service in Lapis",
    license = "MIT",
    homepage = "https://github.com/starius/pasta",
}
dependencies = {
    "lua >= 5.1",
    "lapis",
    "lua-resty-upload",
    "lua-lru",
    "arc4random",
    "lua-mnemonic",
    "luacrypto",
    "lua-filesize",
    "yaml",
}
build = {
    type = "builtin",
    modules = {
        ['pasta.app'] = 'src/pasta/app.lua',
        ['pasta.models'] = 'src/pasta/models.lua',
        ['pasta.view'] = 'src/pasta/view.lua',
        ['pasta.blobs'] = 'src/pasta/blobs.lua',
        ['pasta.views.layout'] = 'src/pasta/views/layout.lua',
        ['pasta.views.view_pasta'] = 'src/pasta/views/view_pasta.lua',
        ['pasta.views.remove'] = 'src/pasta/views/remove.lua',
        ['pasta.views.index'] = 'src/pasta/views/index.lua',
        ['pasta.views.show_password'] =
            'src/pasta/views/show_password.lua',
        ['pasta.views.show_self_burning'] =
            'src/pasta/views/show_self_burning.lua',
        ['pasta.views.edit'] = 'src/pasta/views/edit.lua',
        ['pasta.views.raw_pasta'] = 'src/pasta/views/raw_pasta.lua',
        ['pasta.views.show_url_shortener'] =
            'src/pasta/views/show_url_shortener.lua',
    },
    copy_directories = {
        'src/static',
    },
}
