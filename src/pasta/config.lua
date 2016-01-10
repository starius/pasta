local config = require("lapis.config")

config({"development", "production"}, {
  port = 25516,
})

config("production", {
    code_cache = 'on',
    logging = {queries = false, requests = false},
})
