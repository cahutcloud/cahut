---
title: Run Cahut on your machine
description: Build Cahut from source and serve it over plain HTTP on localhost, to try it without a server.
sidebar:
  label: Run on your machine
---

To try Cahut without a server, build it from source and serve it on your machine over plain HTTP. The releases are for Linux servers only, on amd64 and arm64.

## Build

You need Git, make, Go and Node.js. The `toolchain` line of `go.mod` sets the version of Go, and the `go` command downloads it when yours is older. `.nvmrc` sets the version of Node.js, which builds the web console that the binary embeds.

```bash
git clone https://github.com/getcahut/cahut.git
cd cahut
make build
```

`make build` builds the console, then writes a static binary to `bin/cahut` that holds it. A plain `go build` leaves the console out: such a binary answers its port with a notice.

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

With `--dev`, Cahut serves plain HTTP on `localhost:8080`, without certificates, and the console and its [API](/reference/http-api/) on `localhost:8081`. `--dev-addr` and `--dev-console-addr` change the addresses.

The link that `user create` prints opens the console at `http://localhost:8081/activate#…`, where you choose your password and sign in.

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
bin/cahut token create --data ./cahut-data --domain beta.localhost --user you --name "Claude Code"
```

`http://beta.localhost:8080/` answers as a Cahut of its own, with its own people and tokens: a token of the Cahut at `localhost` does not open it. Add it to Claude Code under another name:

```bash
claude mcp add --transport http cahut-beta http://beta.localhost:8080/mcp --header "Authorization: Bearer <token>"
```

A name that no Cahut declares, such as `nope.localhost`, gets `421 Misdirected Request`, on your machine as on a server. This also guards the MCP endpoints against DNS rebinding: a web page that points its own name at your machine still sends its own name, which no Cahut answers.

## Work on Cahut

In the repository, `make dev` serves the Cahuts of `tmp/dev` on `http://localhost:8080`, and `make` lists the other tasks. `make check` runs the linters and the tests, as CI does.
