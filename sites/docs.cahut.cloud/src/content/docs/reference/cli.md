---
title: "Command line"
description: "Every command of the cahut binary, with its flags, as the binary describes them."
sidebar:
  order: 2
---

<!-- Generated from the code by make docs-reference: do not edit. -->

Cahut is the home for everything your AIs make. The `cahut` binary serves the Cahuts of a server, and administers them from its command line.

```text
cahut <command> [arguments]
```

Run `cahut help` for the list of commands, and add `-h` to a command for its flags. A flag takes one dash or two: `-data` and `--data` are the same. On a server, run the commands as the service user, with the full path of the binary: see [Administration](/admin/commands/).

| Command | What it does |
| --- | --- |
| [`cahut serve`](#cahut-serve) | Serve the Cahuts of this server: MCP for AIs, and the Windows |
| [`cahut domain add`](#cahut-domain-add) | Declare a domain, which creates its Cahut |
| [`cahut domain list`](#cahut-domain-list) | List the Cahuts of this server |
| [`cahut user create`](#cahut-user-create) | Create a person, with their personal Room and an activation link |
| [`cahut user list`](#cahut-user-list) | List the people of a Cahut |
| [`cahut user link`](#cahut-user-link) | Give a person a new link where they choose their password |
| [`cahut token create`](#cahut-token-create) | Create a personal token for an AI |
| [`cahut token list`](#cahut-token-list) | List personal tokens |
| [`cahut token revoke`](#cahut-token-revoke) | Revoke a personal token |
| [`cahut logbook`](#cahut-logbook) | Print the latest entries of the Logbook |
| [`cahut version`](#cahut-version) | Print the version of this binary |

## `cahut serve`

Serve the Cahuts of this server: MCP for AIs, and the Windows.

```text
cahut serve [flags]
```

| Flag | Default | Description |
| --- | --- | --- |
| `--acme-ca string` |  | ACME directory, or "staging" for Let's Encrypt's staging environment (env CAHUT_ACME_CA) |
| `--data string` | `/var/lib/cahut` | directory of the server's data, which holds its Cahuts (env CAHUT_DATA) |
| `--dev` |  | serve plain HTTP on `--dev-addr`, for Cahuts at localhost and \*.localhost |
| `--dev-addr string` | `localhost:8080` | address for `--dev` |
| `--email string` |  | contact of the Let's Encrypt account, optional (env CAHUT_ACME_EMAIL) |
| `--http string` | `:80` | address for HTTP, which answers ACME challenges and redirects |
| `--https string` | `:443` | address for HTTPS |

## `cahut domain`

Manage the Cahuts of this server, one per domain.

### `cahut domain add`

Declare a domain, which creates its Cahut.

```text
cahut domain add [--data DIR] <domain>
```

| Flag | Default | Description |
| --- | --- | --- |
| `--data string` | `/var/lib/cahut` | directory of the server's data, which holds its Cahuts (env CAHUT_DATA) |

### `cahut domain list`

List the Cahuts of this server.

```text
cahut domain list [flags]
```

| Flag | Default | Description |
| --- | --- | --- |
| `--data string` | `/var/lib/cahut` | directory of the server's data, which holds its Cahuts (env CAHUT_DATA) |

## `cahut user`

Manage the people of a Cahut.

### `cahut user create`

Create a person, with their personal Room and an activation link.

```text
cahut user create [flags]
```

| Flag | Default | Description |
| --- | --- | --- |
| `--admin` |  | make the person an administrator of the Cahut |
| `--data string` | `/var/lib/cahut` | directory of the server's data, which holds its Cahuts (env CAHUT_DATA) |
| `--domain string` |  | the Cahut to act in, by its domain; optional when the server holds a single Cahut |
| `--handle string` |  | short name, which also names the personal Room: 2 to 32 lowercase letters, digits and hyphens |
| `--name string` |  | full name |

### `cahut user list`

List the people of a Cahut.

```text
cahut user list [flags]
```

| Flag | Default | Description |
| --- | --- | --- |
| `--data string` | `/var/lib/cahut` | directory of the server's data, which holds its Cahuts (env CAHUT_DATA) |
| `--domain string` |  | the Cahut to act in, by its domain; optional when the server holds a single Cahut |

### `cahut user link`

Give a person a new link where they choose their password.

```text
cahut user link [flags]
```

| Flag | Default | Description |
| --- | --- | --- |
| `--data string` | `/var/lib/cahut` | directory of the server's data, which holds its Cahuts (env CAHUT_DATA) |
| `--domain string` |  | the Cahut to act in, by its domain; optional when the server holds a single Cahut |
| `--handle string` |  | the person who chooses a new password |

## `cahut token`

Manage the personal tokens of AIs.

### `cahut token create`

Create a personal token for an AI.

```text
cahut token create [flags]
```

| Flag | Default | Description |
| --- | --- | --- |
| `--data string` | `/var/lib/cahut` | directory of the server's data, which holds its Cahuts (env CAHUT_DATA) |
| `--domain string` |  | the Cahut to act in, by its domain; optional when the server holds a single Cahut |
| `--name string` |  | what the token is for, such as "Claude Code on my laptop" |
| `--user string` |  | handle of the person the AI acts for |

### `cahut token list`

List personal tokens.

```text
cahut token list [flags]
```

| Flag | Default | Description |
| --- | --- | --- |
| `--data string` | `/var/lib/cahut` | directory of the server's data, which holds its Cahuts (env CAHUT_DATA) |
| `--domain string` |  | the Cahut to act in, by its domain; optional when the server holds a single Cahut |
| `--user string` |  | only the tokens of this person |

### `cahut token revoke`

Revoke a personal token.

```text
cahut token revoke [flags]
```

| Flag | Default | Description |
| --- | --- | --- |
| `--data string` | `/var/lib/cahut` | directory of the server's data, which holds its Cahuts (env CAHUT_DATA) |
| `--domain string` |  | the Cahut to act in, by its domain; optional when the server holds a single Cahut |
| `--id int` |  | the token to revoke, as token list shows it |

## `cahut logbook`

Print the latest entries of the Logbook.

```text
cahut logbook [flags]
```

| Flag | Default | Description |
| --- | --- | --- |
| `--data string` | `/var/lib/cahut` | directory of the server's data, which holds its Cahuts (env CAHUT_DATA) |
| `--domain string` |  | the Cahut to act in, by its domain; optional when the server holds a single Cahut |
| `--limit int` | `20` | how many entries |
| `--server` |  | print the Logbook of the server, not of a Cahut |

## `cahut version`

Print the version of this binary.

```text
cahut version
```

This command takes no flags.
