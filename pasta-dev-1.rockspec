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
}
build = {
    type = "builtin",
    modules = {
        ['pasta'] = 'src/pasta/pasta.lua',
    },
}
