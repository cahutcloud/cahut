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
| [`cahut setup`](#cahut-setup) | Give the link to the first start of a Cahut, in the browser |
| [`cahut domain add`](#cahut-domain-add) | Declare a domain, which creates its Cahut |
| [`cahut domain list`](#cahut-domain-list) | List the Cahuts of this server |
| [`cahut domain rename`](#cahut-domain-rename) | Move a Cahut to another domain, or from the IP address of the server to a domain |
| [`cahut settings show`](#cahut-settings-show) | Print the kind and limits of a Cahut, with what it holds |
| [`cahut settings set`](#cahut-settings-set) | Change the kind or limits of a Cahut; what you leave out stays as it is |
| [`cahut admin create`](#cahut-admin-create) | Create an administrator of the server, with an activation link |
| [`cahut admin list`](#cahut-admin-list) | List the administrators of the server |
| [`cahut admin link`](#cahut-admin-link) | Give an administrator of the server a new link where they choose their password |
| [`cahut user create`](#cahut-user-create) | Create a person, with their personal Room and an activation link |
| [`cahut user list`](#cahut-user-list) | List the people of a Cahut |
| [`cahut user link`](#cahut-user-link) | Give a person a new link where they choose their password |
| [`cahut token create`](#cahut-token-create) | Create a personal token for an AI |
| [`cahut token list`](#cahut-token-list) | List personal tokens |
| [`cahut token revoke`](#cahut-token-revoke) | Revoke a personal token |
| [`cahut logbook`](#cahut-logbook) | Print the latest entries of the Logbook |
| [`cahut update`](#cahut-update) | Install a new release, as root, and go back if it fails |
| [`cahut doctor`](#cahut-doctor) | Diagnose this server, and say what to fix |
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
| `--update-state string` | `/var/lib/cahut-update` | directory where the updater leaves the state of updates (env CAHUT_UPDATE_STATE) |

## `cahut setup`

Give the link to the first start of a Cahut, in the browser.

```text
cahut setup [flags]
```

| Flag | Default | Description |
| --- | --- | --- |
| `--data string` | `/var/lib/cahut` | directory of the server's data, which holds its Cahuts (env CAHUT_DATA) |
| `--domain string` |  | the Cahut to act in, by its domain; optional when the server holds a single Cahut |

## `cahut domain`

Manage the Cahuts of this server, one per domain.

### `cahut domain add`

Declare a domain, which creates its Cahut.

```text
cahut domain add [--data DIR] [--admin-handle HANDLE --admin-name NAME] <domain>
```

| Flag | Default | Description |
| --- | --- | --- |
| `--admin-handle string` |  | handle of the first administrator of the Cahut, a person of it, who gets an activation link |
| `--admin-name string` |  | full name of the first administrator |
| `--data string` | `/var/lib/cahut` | directory of the server's data, which holds its Cahuts (env CAHUT_DATA) |

### `cahut domain list`

List the Cahuts of this server.

```text
cahut domain list [flags]
```

| Flag | Default | Description |
| --- | --- | --- |
| `--data string` | `/var/lib/cahut` | directory of the server's data, which holds its Cahuts (env CAHUT_DATA) |

### `cahut domain rename`

Move a Cahut to another domain, or from the IP address of the server to a domain.

```text
cahut domain rename [--data DIR] <domain> <new domain>
```

| Flag | Default | Description |
| --- | --- | --- |
| `--data string` | `/var/lib/cahut` | directory of the server's data, which holds its Cahuts (env CAHUT_DATA) |

## `cahut settings`

Set the kind and limits of a Cahut.

### `cahut settings show`

Print the kind and limits of a Cahut, with what it holds.

```text
cahut settings show [flags]
```

| Flag | Default | Description |
| --- | --- | --- |
| `--data string` | `/var/lib/cahut` | directory of the server's data, which holds its Cahuts (env CAHUT_DATA) |
| `--domain string` |  | the Cahut to act in, by its domain; optional when the server holds a single Cahut |

### `cahut settings set`

Change the kind or limits of a Cahut; what you leave out stays as it is.

```text
cahut settings set [--data DIR] [--domain DOMAIN] [--kind KIND] [--people N] [--admins N] [--rooms N] [--windows N] [--windows-per-room N] [--disk SIZE] [--rooms-per-person N] [--windows-per-person N] [--disk-per-person SIZE]
```

| Flag | Default | Description |
| --- | --- | --- |
| `--admins N` |  | N of them may be administrators of the Cahut; none lifts the limit |
| `--data string` | `/var/lib/cahut` | directory of the server's data, which holds its Cahuts (env CAHUT_DATA) |
| `--disk SIZE` |  | the Cahut may keep SIZE, such as 10GB or 500MB, in units of 1000 bytes; none lifts the limit |
| `--disk-per-person SIZE` |  | in an individual Cahut, the Windows of each person may keep SIZE; none lifts the limit |
| `--domain string` |  | the Cahut to act in, by its domain; optional when the server holds a single Cahut |
| `--kind string` |  | collaborative or individual; the kind stays when left out |
| `--people N` |  | the Cahut may have N people, administrators included, and disabled people do not count; none lifts the limit |
| `--rooms N` |  | the Cahut may have N team Rooms; none lifts the limit |
| `--rooms-per-person N` |  | in an individual Cahut, each person may have N Rooms beside their personal Room; none lifts the limit |
| `--windows N` |  | the Cahut may hold N Windows out of the trash; none lifts the limit |
| `--windows-per-person N` |  | in an individual Cahut, each person may hold N Windows out of the trash; none lifts the limit |
| `--windows-per-room N` |  | each Room may hold N Windows out of the trash; none lifts the limit |

## `cahut admin`

Manage the administrators of the server.

### `cahut admin create`

Create an administrator of the server, with an activation link.

```text
cahut admin create [flags]
```

| Flag | Default | Description |
| --- | --- | --- |
| `--data string` | `/var/lib/cahut` | directory of the server's data, which holds its Cahuts (env CAHUT_DATA) |
| `--domain string` |  | the Cahut whose console the link opens; the first one declared when empty |
| `--handle string` |  | short name of the administrator: 2 to 32 lowercase letters, digits and hyphens |
| `--name string` |  | full name |

### `cahut admin list`

List the administrators of the server.

```text
cahut admin list [flags]
```

| Flag | Default | Description |
| --- | --- | --- |
| `--data string` | `/var/lib/cahut` | directory of the server's data, which holds its Cahuts (env CAHUT_DATA) |

### `cahut admin link`

Give an administrator of the server a new link where they choose their password.

```text
cahut admin link [flags]
```

| Flag | Default | Description |
| --- | --- | --- |
| `--data string` | `/var/lib/cahut` | directory of the server's data, which holds its Cahuts (env CAHUT_DATA) |
| `--domain string` |  | the Cahut whose console the link opens; the first one declared when empty |
| `--handle string` |  | the administrator who chooses a new password |

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

## `cahut update`

Install a new release, as root, and go back if it fails.

```text
sudo cahut update [--check] [--version VERSION] [--binary FILE]
```

| Flag | Default | Description |
| --- | --- | --- |
| `--bin string` | `/usr/local/bin/cahut` | the binary that the service runs |
| `--binary string` |  | install this cahut binary instead of a release, with the same check and rollback, without signature |
| `--check` |  | read the releases and say whether a newer one is out, without installing it |
| `--data string` | `/var/lib/cahut` | directory of the server's data, which holds its Cahuts (env CAHUT_DATA) |
| `--state string` | `/var/lib/cahut-update` | directory of the state of updates |
| `--unattended` |  | run as systemd does: what the console asked for, or else what CAHUT_UPDATES says |
| `--version string` |  | install this release, such as 0.4.1, instead of the newest; never an older one |

## `cahut doctor`

Diagnose this server, and say what to fix.

```text
sudo cahut doctor [--data DIR]
```

| Flag | Default | Description |
| --- | --- | --- |
| `--bin string` | `/usr/local/bin/cahut` | the binary that the service runs |
| `--data string` | `/var/lib/cahut` | directory of the server's data, which holds its Cahuts (env CAHUT_DATA) |
| `--state string` | `/var/lib/cahut-update` | directory of the state of updates |

## `cahut version`

Print the version of this binary.

```text
cahut version
```

This command takes no flags.
