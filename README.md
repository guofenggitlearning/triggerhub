# Trigger Hub

Trigger Hub is a simple service that listens for trigger events on HTTP clients and relays them to subscribed services that may not want to expose a dedicated port. In this way, triggers are pulled on the service instead of directly pushed by clients, which adds an extra layer of isolation.

## Get started

[Download the Linux binary](https://github.com/brickpop/triggerhub/releases/download/v0.1.0/triggerhub.zip) or compile the project yourself:

```sh
go build -o build/triggerhub
```

## Config

Parameters can be passed via command line of be defined in a config file

### Command line

```
$ ./triggerhub -h
Trigger Hub is a simple service that listens for trigger events on HTTP clients and relays them to subscribed services that may not want to expose a dedicated port.

Usage:
  triggerhub [command]

Available Commands:
  help        Help about any command
  join        Joins a Trigger Hub server
  serve       Start a dispatcher service

Flags:
      --config string   the config file to use
  -h, --help            help for triggerhub

Use "triggerhub [command] --help" for more information about a command.
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
$ triggerhub --config my-config.yaml
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

Then, run `triggerhub`:

```sh
$ triggerhub --tls --cert cert.pem --key key.pem -p 8443
```

And finally, tell `curl` to accept the self-signed certificate.

```sh
$ NAME = "media"
$ TOKEN = "1234567890123456789012345678901234567890"
$ curl --insecure https://localhost:8443/backup/$NAME/$TOKEN > $NAME.tar.gz
```
