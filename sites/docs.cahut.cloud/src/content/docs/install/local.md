---
title: Run Cahut on your machine
description: Build Cahut from source and serve it over plain HTTP on localhost, to try it without a server.
sidebar:
  label: Run on your machine
---

To try Cahut without a server, build it from source and serve it on your machine over plain HTTP. The releases are for Linux servers only, on amd64 and arm64.

## Build

You need Git, make and Go. The `toolchain` line of `go.mod` sets the version of Go, and the `go` command downloads it when yours is older.

```bash
git clone https://github.com/getcahut/cahut.git
cd cahut
make build
```

`make build` writes a static binary to `bin/cahut`.

## Declare a Cahut at localhost

Keep the data in a folder of your choice with `--data`:

```bash
bin/cahut domain add --data ./cahut-data localhost
bin/cahut user create --data ./cahut-data --handle you --name "Your Name" --admin
bin/cahut token create --data ./cahut-data --user you --name "Claude Code"
```

The last command prints the token once: copy it.

## Serve

```bash
bin/cahut serve --dev --data ./cahut-data
```

With `--dev`, Cahut serves plain HTTP on `localhost:8080`, without certificates. `--dev-addr` changes the address.

Add the Cahut to Claude Code, with the token:

```bash
claude mcp add --transport http cahut-dev http://localhost:8080/mcp --header "Authorization: Bearer <token>"
```

Windows then open at `http://localhost:8080/<room>/<window>/`.

## A second Cahut

Browsers and curl send `localhost` and every name under it, such as `beta.localhost`, to your machine. Declare `beta.localhost` to see how one server holds two Cahuts. Once the server holds more than one, the commands that act in a Cahut take `--domain`:

```bash
bin/cahut domain add --data ./cahut-data beta.localhost
bin/cahut user create --data ./cahut-data --domain beta.localhost --handle you --name "Your Name"
```

`http://beta.localhost:8080/` answers as a Cahut of its own. Its MCP endpoint, though, turns AIs away on your machine: on a local address, the MCP endpoint accepts the host name `localhost` only, as a guard against DNS rebinding. With `--dev`, AIs publish to the Cahut at `localhost`.

## Work on Cahut

In the repository, `make dev` serves the Cahuts of `tmp/dev` on `http://localhost:8080`, and `make` lists the other tasks. `make check` runs the linters and the tests, as CI does.
