# pasta, Pastebin-like service in Lapis

## How to use

1. Install [Lapis](http://leafo.net/lapis)
2. Install [OpenResty](https://openresty.org/)
3. Install [LuaRocks](https://luarocks.org/)
4. Setup the database and change `pasta/src/config.lua`
5. `cd pasta && sudo luarocks make`
6. `cd pasta/src && lapis migrate`
7. `cd pasta/src && lapis server`

If your Nginx has multiple workers, disable cache in `config.lua`.

You can also use it without OpenResty, but in Nginx with LuaJIT support.
In Debian it can be installed with package `nginx-extras`.
Copy file `src/nginx.conf.compiled` generated by `lapis server production`
to `/etc/nginx/sites-available/pasta` and add a symlink to it in
`/etc/nginx/sites-enabled`. Change it for your setup and add the following
line after `content_by_lua`:

```lua
package.loaded.config = dofile "/path/to/your/config.lua"
```

See also Nginx configuration used for site pasta.cf:
[src/nginx-production][1].

[1]: https://github.com/starius/pasta/tree/master/src/nginx-production
