# 多传感器融合同步 误差5ms
# Trigger Hub

Trigger Hub is a simple service that listens for trigger events on HTTP clients and relays them to subscribed services that may not want to expose a dedicated port. In this way, triggers are pulled on the service instead of directly pushed by clients, which adds an extra layer of isolation.

## Get started

[Download the Linux binary](https://github.com/brickpop/triggerhub/releases/download/v0.1.0/triggerhub.zip) or compile the project yourself:

```sh
go build -o build/triggerhub
```

## Parameters

Parameters can be passed via command line of be defined in a config file

### Command line

```
$ ./triggerhub -h
Trigger Hub is a simple service that listens for trigger events on HTTP clients and relays them to subscribed services that may not want to expose a dedicated port.

Usage:
  triggerhub [command]

Available Commands:
  help        Help about any command
  listen      Joins a Trigger Hub server
  serve       Start a dispatcher service

Flags:
      --config string   the config file to use
  -h, --help            help for triggerhub

Use "triggerhub [command] --help" for more information about a command.
```

## Components

### Dispatcher

The dispatcher is typically running on an external server (different to the target server handling actual work load). The dispatcher exposes an HTTP(s)/WS(s) endpoint that can be called by two actors:

- Web clients
  - Typically from a web browser
  - Requesting the invokation of an action on the target server
- The target server (listener)
  - Declares the supported action ID's that dispatcher clients may call
  - Listens to messages on the dispatcher's web socket

#### Config file

In dispatcher mode, the config file is optional. By default, triggerhub attempts to use `config.yaml` from the current directory. To use a config file on a different path:

```sh
$ triggerhub serve --config my-config.yaml
```

Copy `config-dispatcher.template.yaml` and adapt it to work for the data you want to back up.

```yaml
token: "1234"    # Authentication token for target servers
port: 8443
tls: true
cert: /path/to/cert.pem   # not needed it tls = false
key: /path/to/key.pem     # not needed it tls = false
```

CLI parameters can override the `port`, `cert`, `key` and `tls` for easier testing. 

### Listener

Trigger Hub listeners are the hosts where the actual action is going to be executed. When running as a listener, no port is exposed and rather, a WS connection is made to the dispatcher. 

#### Config file

The config file is required to run in listener mode. By default, triggerhub also tries to use `config.yaml` from the current directory. To use a config file on a different path:

```sh
$ triggerhub listen --config my-config.yaml
```

Copy `config-listener.template.yaml` and adapt it to work for the data you want to back up.

```yaml
name: my-listener-name
dispatcher:
  url: wss://dispatcher-domain.com/ws
  token: "1234"   # Has to match the server's auth token
triggers:
- action: restart-service-1
  token: "1234567890123456789012345678901234567890"
  command: /home/user/scripts/restart.sh
- action: restart-service-2
  token: "0123456789012345678901234567890123456789"
  command: /home/user/scripts/restart.sh
```
