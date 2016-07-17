local db = require("lapis.db")
local schema = require("lapis.db.schema")

local migrations = {}

function migrations.rev_2016_01_16_init()
    schema.create_table("pasta", {
        {"hash", schema.types.varchar},
        {"self_burning", schema.types.boolean},
        {"filename", schema.types.varchar},
        {"salt", schema.types.varchar},
        {"content", schema.types.text},
        {"password", schema.types.varchar},
    })
    schema.create_index("pasta", "hash")
end

function migrations.rev_2016_02_19_init()
    db.query("ALTER TABLE pasta ADD PRIMARY KEY (hash);")
end

function migrations.rev_2016_07_17_redirect_to()
    schema.add_column("pasta", "redirect_to", schema.types.boolean)
end

return migrations
