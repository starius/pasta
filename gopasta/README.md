# gopasta - pastebin-like site as single Go binary

Public instance: https://pastacity.nl/

Onion site mirror: http://trebzwp2nyrsq6urxkeif7iz5vpf3ppnlzvy6xdxwomgebx2vpczzsqd.onion/

Below are instructions on how to run your own instance.

## Build from source

Install Go: https://go.dev/doc/install

You need at least Go 1.27.3. Use the latest stable patch release for security fixes.

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

## Run it on server

First generate encryption password:

```
./gopasta -gen-secret
liar hunter dotted wedge hydrogen pegs purged suffice opposite smuggled paddles folding
```

(The generated secret will be different.)

Write down the secret. The binary will ask it on startup.
If you forget it, the data can not be recovered.
Use the generated phrase, not words you choose yourself. Empty phrases, malformed
spacing, and phrases encoding fewer than 16 bytes are rejected. Do not change the
secret for an existing database: its files and links depend on that secret.

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
 - `-listen` defaults to `127.0.0.1:8042`; expose this plaintext listener only on a trusted network
 - `-max-requests` limits simultaneous application requests across HTTP and HTTPS (default 16); excess requests receive HTTP 503
 - `-read-timeout` and `-write-timeout` limit request and response duration (default `2m` each); increase them if large uploads over slow connections require it

You can get error like "listen tcp :443: bind: permission denied".
Run the daemon as a dedicated, unprivileged user with a private state directory
(mode `0700`). To allow that user to bind ports 443 and 80, grant the binary the
capability to bind privileged ports:

```
sudo apt-get install libcap2-bin
sudo setcap 'cap_net_bind_service=+ep' /path/to/gopasta
```

Alternatively, put a TLS reverse proxy in front of the loopback listener.

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

## Security properties and limits

Secure IDs are selected by default in the browser form. API clients should send
`long_id=on` for private pastes. Short word links are intentionally guessable;
the first 2,048 records have only one word each. Anyone who knows a link can read
it. Self-burning removes the server's copy after one successful lookup; it cannot
erase copies already downloaded by a recipient or retained in backups and
filesystem snapshots. Responses instruct browsers and
proxies not to cache pastes or send their URLs in referrers.

Uploaded active content such as HTML, SVG, and JavaScript is downloaded as an
attachment. Uploaded-content responses also use a restrictive CSP sandbox.
Shortened URLs must use HTTP or HTTPS.

The existing on-disk format uses **unauthenticated AES-CTR**. It does not detect
malicious modification of the database files, and reusing a secret for a fresh
database or writing new data after rolling back a backup can reuse the encryption
keystream. Use a fresh generated secret for each new database and protect files
and backups from other users. Fully addressing these limitations requires a
versioned authenticated storage format and migration of existing data; the
current hardening preserves the existing format.
Run only one daemon per state directory: the format has no inter-process locking
or transactional crash recovery. A storage write error disables further uploads
until restart; fix the storage problem and verify the database before restarting.

Anonymous uploads can still fill the disk over time. Set a filesystem quota and
monitor free space, and apply rate limits at a reverse proxy when exposing a
public instance. Tune `-max-requests`, `-max-size`, and cache limits together to
fit available memory. Transmit the administrator Authorization token only over
HTTPS, an onion service, or a trusted local connection.
