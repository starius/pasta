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
    "lua-lru",
    "arc4random",
    "lua-mnemonic",
    "lxsh",
    "luacrypto",
    "filesize",
}
build = {
    type = "builtin",
    modules = {
        ['pasta.app'] = 'src/pasta/app.lua',
        ['pasta.views.index'] = 'src/pasta/views/index.lua',
        ['pasta.models'] = 'src/pasta/models.lua',
    },
}
