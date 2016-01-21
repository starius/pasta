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
    nwords_short = 3,
    nwords_long = 9,
    max_pasta_size = 10000000,
    max_filename = 150,
    cache_nrecords = 1000,
    cache_nbytes = 100000000,
})

config("production", {
    code_cache = 'on',
    logging = {queries = false, requests = false},
    print_stack_to_browser = false,
})

config("development", {
    print_stack_to_browser = true,
})
