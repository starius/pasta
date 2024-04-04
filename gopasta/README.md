# gopasta - pastebin-like site as single Go binary

Public instance: https://pastacity.nl/

Onion site mirror: http://trebzwp2nyrsq6urxkeif7iz5vpf3ppnlzvy6xdxwomgebx2vpczzsqd.onion/

Below are instructions on how to run your own instance.

## Build from source

Install Go: https://go.dev/doc/install

You need at least version 1.19.

To download the code and build the binary, use the following command:
```
$ go install github.com/starius/pasta/gopasta@latest
```

Some optional modifications:

 - add `-ldflags="-w -s"` after `install` to reduce the size of the binary
 - add `-trimpath` and `-ldflags="-buildid="` to reduce make the binary build reproducible
 - set `CGO_ENABLED=0` environment variable to make the binary static

It will put the binary to `$GOPATH/bin` directory, which is in most setups in `~/go/bin` directory.

Alternatively you can checkout the repo and build binary in the current dir:

```
$ git clone https://github.com/starius/pasta
$ cd pasta/gopasta
$ CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -buildid=" -o gopasta
```

Run `./release.sh` to produce release binaries in the current directory.
They should match the ones in https://github.com/starius/pasta/releases
for the given Go version and this package version.

The binary is self-contained, so if needed you can move it to another machine
with the same OS and CPU architecture and run there.

### Using C++ RE2

If you send a one time link in some chat (e.g. discord), its crawler will try to
load link preview and would [expire the link][issue8], if not the protection
that we have. The crawlers are detected and blocked in one-time links, so they
do not expire. Regular expressions matching against all the crawlers User Agent
strings is done using [go-re2](https://github.com/wasilibs/go-re2) library which
is very fast. To make this matching even faster (~3 times), you can build
against C++ version of RE2 library:

```
sudo apt-get install libre2-dev
sudo apt-get install build-essential pkg-config
go build -trimpath -ldflags="-s -w -buildid=" -tags re2_cgo -o gopasta
```

The library is linked dynamically, so it has to be installed on the system
on which it is run.

[issue8]: https://github.com/starius/pasta/issues/8

## Run it on server

First generate encryption password:

```
./gopasta -gen-secret
liar hunter dotted wedge hydrogen pegs purged suffice opposite smuggled paddles folding
```

(The generated secret will be different.)

Write down the secret. The binary will ask it on startup.
If you forget it, the data can not be recovered.

Now generate the command to delete links (as admin):

```
./gopasta -print-admin-auth
Enter secret:
<enter the secret>
curl -X DELETE -H 'Authorization: 482b2ed338e8107d5a0010d90ecbfb21' URL-to-delete
```

Write down the template of curl command and keep it secretly.
In case you need to quickly remove a record, replace `URL-to-delete` with
actual URL of the record and run the command.

Now let's run the site!

To run the binary:

```
./gopasta \
  -cache-bytes 0 -cache-records 0 \
  -allow-files -files-burn \
  -dir /path/to/directory/with/state \
  -listen 127.0.0.1:8042 \
  -max-size 104857600 \
  -domains pastacity.nl,trebzwp2nyrsq6urxkeif7iz5vpf3ppnlzvy6xdxwomgebx2vpczzsqd.onion \
  -letsencrypt-domains pastacf.com,www.pastacf.com,pastacity.nl,www.pastacity.nl
Enter secret:
<enter the secret>
```

Adjust the flags:

 - `-dir` is the directory where the uploads are stored in encrypted form
 - `-domains` is the list of domains to show in the bottom of main page
 - `-letsencrypt-domains` is the list of allowed domains to issue letsencrypt cert for it automatically
 - remove flag `-allow-files` to disable file uploads
 - remove `-files-burn` to allow permanent links to file uploads
 - adjust `-max-size` to change maximum upload size (in bytes)

You can get error like "listen tcp :443: bind: permission denied".
The binary needs to bind 443 and 80 TCP ports on the server to run HTTP server,
so run as root or run the following command to allow the binary to bind privileged ports (below 1024):

```
sudo apt-get install libcap2-bin
sudo setcap 'cap_net_bind_service=+ep' /path/to/gopasta
```

If this doesn't work, allow any binary to bind any port >= 80 using this command:
```
sudo sysctl -w net.ipv4.ip_unprivileged_port_start=80
```

The binary will run the site on the following ports:

 - https://0.0.0.0:443 - externally available HTTPS site
 - http://0.0.0.0:80 - externally available HTTP server, redirecting to HTTPS site
 - http://127.0.0.1:8042 - the site available locally and used by onion site

To add onion mirror, [create an onion site](https://community.torproject.org/onion-services/setup/)
in torrc config and point its 80 port to 127.0.0.1:8042:
```
HiddenServiceDir /var/lib/tor/hidden_service/
HiddenServicePort 80 127.0.0.1:8042
```
