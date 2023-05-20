# gopasta - pastebin-like site as single Go binary

Public instance: https://pastacity.nl/

Onion site mirror: http://trebzwp2nyrsq6urxkeif7iz5vpf3ppnlzvy6xdxwomgebx2vpczzsqd.onion/

Below are instructions on how to run your own instance.

## Build from source

Install Go: https://go.dev/doc/install

You need at least version 1.17.

To download the code and build the binary, use the following command:
```
$ go install github.com/starius/pasta/gopasta@latest
```

Some optional modifications:

 - add `-ldflags="-w -s"` after `install` to reduce the size of the binary
 - set `CGO_ENABLED=0` environment variable to make the binary static

It will put the binary to `$GOPATH/bin` directory, which is in most setups in `~/go/bin` directory.

The binary is self-contained, so if needed you can move it to another machine
with the same OS and CPU architecture and run there.

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
In case you need quickly remove a record, replace `URL-to-delete` with
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

To add onion mirror, [create an onion site](https://community.torproject.org/onion-services/setup/)
in torrc config and point its 80 port to 127.0.0.1:8042:
```
HiddenServiceDir /var/lib/tor/hidden_service/
HiddenServicePort 80 127.0.0.1:8042
```
