# Packerd

Packerd is a simple service that generates on-demand backup bundles and serves them remotely to authorized clients.

## Get started

[Download the Linux binary](https://github.com/brickpop/packerd/releases/download/v0.1.0/packerd.zip) or compile the project yourself:

```sh
go build -o build/packerd
```

## Config

Parameters can be passed via command line of be defined in a config file

### Command line

```
$ ./packerd -h
Packerd is a backup utility for authenticated remote clients

Usage:
  packerd [flags]

Flags:
      --cert string     the certificate file (TLS only)
      --config string   the config file to use
  -h, --help            help for packerd
      --key string      the TLS encryption key file
  -p, --port int        port to bind to (default 80)
      --tls             whether to use TLS encryption (cert and key required)
```

The config file is mandatory. However, CLI parameters can override the `port`, `cert`, `key` and `tls` for easier testing. 

### Config file

Copy `config.template.yaml` and adapt it to work for the data you want to back up.


```yaml
port: 8443
tls: true
cert: /path/to/cert.pem   # not needed it tls = false
key: /path/to/key.pem     # not needed it tls = false
paths:
- id: media
  token: "1234567890123456789012345678901234567890"
  path: /var/lib/media
- id: other
  token: "0123456789012345678901234567890123456789"
  path: /var/lib/other
```

By default, `config.yaml` is attempted on the current directory. Otherwise:

```sh
$ packerd --config my-config.yaml
```

## Fetching the back up remotely

From a remote location, make a GET request like so:

```sh
$ NAME = "media"
$ TOKEN = "1234567890123456789012345678901234567890"
$ curl https://my-server.net:8443/backup/$NAME/$TOKEN > $NAME.tar.gz
```

## Using a self signed certificate

```sh
$ openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 10000 -nodes
```

Then, run `packerd`:

```sh
$ packerd --tls --cert cert.pem --key key.pem -p 8443
```

And finally, tell `curl` to accept the self-signed certificate.

```sh
$ NAME = "media"
$ TOKEN = "1234567890123456789012345678901234567890"
$ curl --insecure https://localhost:8443/backup/$NAME/$TOKEN > $NAME.tar.gz
```
