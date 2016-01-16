local config = require("lapis.config")

config({"development", "production"}, {
    port = 25516,
    postgres = {
        backend = "pgmoon",
        host = "127.0.0.1",
        user = "myUser",
        password = "myPassword",
        database = "myDatabase",
    },
})

config("production", {
    code_cache = 'on',
    logging = {queries = false, requests = false},
})
