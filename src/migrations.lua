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

return migrations
