lua_code_cache on;
client_max_body_size 10485760;
client_body_buffer_size 10485760;

location / {
    access_log off;
    error_log /dev/null crit;
    default_type text/html;
    content_by_lua '
        package.loaded.config =
            dofile "/etc/nginx/sites-available/pasta-config.lua"
        require("lapis").serve("pasta.app")
    ';
}
