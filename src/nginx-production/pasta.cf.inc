lua_code_cache on;
client_max_body_size 10485760;
client_body_buffer_size 10485760;

location /static/ {
    expires 365d;
    alias /usr/local/lib/luarocks/rocks/pasta/dev-1/src/static/;
}

location /favicon.ico {
    alias static/pasta/favicon.ico;
}

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
