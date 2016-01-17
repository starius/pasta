local config = require("lapis.config")

config({"development", "production"}, {
    port = 25516,
    postgres = {
        backend = "pgmoon",
        host = "127.0.0.1",
        user = "foo",
        password = "bar",
        database = "foodb",
    },
    hash_secret1 = 'hash_secret1',
    hash_secret2 = 'hash_secret2',
})

config("production", {
    code_cache = 'on',
    logging = {queries = false, requests = false},
})
