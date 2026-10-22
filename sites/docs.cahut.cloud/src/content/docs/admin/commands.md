---
title: Administer from the command line
description: Declare Cahuts, create people and tokens, and read the Logbook with the cahut command, on the server.
sidebar:
  label: Command line
---

Until the console comes, you administer Cahut on the server, with the `cahut` command. Each command runs an operation of the registry as the administrator, and its changes go to the Logbook.

## Run the commands

Run them as the `cahut` user, which owns the data, and give the full path of the binary:

```bash
sudo -u cahut /usr/local/bin/cahut domain list
```

On AlmaLinux, `sudo` does not search `/usr/local/bin`, hence the full path. The commands read the data in `/var/lib/cahut`: `--data`, or the variable `CAHUT_DATA`, points them to another folder.

When the server holds several Cahuts, the commands that act in one of them need `--domain`. With a single Cahut, they act in it.

The [command line reference](/reference/cli/) lists every command and flag.

## Declare a Cahut

Point the DNS record of the domain to the server first, then:

```bash
sudo -u cahut /usr/local/bin/cahut domain add docs.example.org
sudo -u cahut /usr/local/bin/cahut domain list
```

```text
Declared docs.example.org: its Cahut is ready. Point its DNS to this server, then serve it.
ID  DOMAIN             CREATED
1   cahut.example.com  2026-10-20
2   docs.example.org   2026-10-21
```

The running service serves the new Cahut at once, and gets its certificate at the first HTTPS connection to its domain. A domain is a host name such as `cahut.example.com`: IP addresses are not supported. There is no command yet to remove a Cahut or to change its domain.

`--admin-handle` and `--admin-name` create the first administrator of the new Cahut at the same time, a person of it, and print their activation link, as [`user create`](#create-a-person) does:

```bash
sudo -u cahut /usr/local/bin/cahut domain add docs.example.org --admin-handle ana --admin-name "Ana Lima"
```

## Create a person

```bash
sudo -u cahut /usr/local/bin/cahut user create --domain cahut.example.com --handle ana --name "Ana Lima"
```

```text
Created ana (Ana Lima), member of cahut.example.com, with a personal Room at cahut.example.com/ana.

Send ana this link, where they choose their password. It works once, until 2026-10-29, and is shown only now.

https://cahut.example.com:8443/activate#...
```

The handle names the person and their personal Room: 2 to 32 lowercase letters, digits and hyphens, starting with a letter and ending with a letter or a digit. It must be free, as the name of a person and of a Room, and cannot be `mcp`, `api`, `admin`, `console` or `help`. `--admin` makes the person an administrator of the Cahut.

The link opens the console of the Cahut, on port 8443, where the person chooses a password of 15 characters at least. It works once, for 7 days. The console is not available yet: until it comes, people act through their AIs, with [personal tokens](/connect/tokens/).

## List people and give a new link

```bash
sudo -u cahut /usr/local/bin/cahut user list --domain cahut.example.com
sudo -u cahut /usr/local/bin/cahut user link --domain cahut.example.com --handle ana
```

```text
HANDLE   NAME            ROLE    STATUS
ana      Ana Lima        member  pending
olivier  Olivier Durand  admin   active
```

A person is `pending` until they choose a password, then `active`, or `disabled`. `user link` gives a new link to a person who lost theirs or forgot their password: it replaces their previous link, and their password, if they have one, works until they use the new one. It also works for administrators, so that the command line is the way back in when nobody can sign in.

Changing a person's role, and disabling or enabling a person, are operations for the administrators of the Cahut, which the console will offer.

## Administrators of the server

The administrators of the server have accounts of their own, apart from the people of the Cahuts. In the console, they declare Cahuts and read the Logbook of the server, and in each Cahut they manage its people and its Rooms, but do not open its Windows. They sign in at the console of any Cahut: a session holds at the domain where it opened, so an administrator who looks after several Cahuts signs in to each.

```bash
sudo -u cahut /usr/local/bin/cahut admin create --handle olivier --name "Olivier Durand"
sudo -u cahut /usr/local/bin/cahut admin list
sudo -u cahut /usr/local/bin/cahut admin link --handle olivier
```

```text
Created olivier (Olivier Durand), administrator of this server.

Send olivier this link, where they choose their password. It works once, until 2026-10-29, and is shown only now.

https://cahut.example.com:8443/server/activate#...
```

The link opens at the console of the first Cahut declared, or of the one `--domain` names. `admin link` gives a new link to an administrator who lost theirs or forgot their password. The command line itself needs no account: whoever runs it on the server acts as an administrator of the server, everywhere.

## Manage tokens

```bash
sudo -u cahut /usr/local/bin/cahut token create --domain cahut.example.com --user ana --name "Claude Code"
sudo -u cahut /usr/local/bin/cahut token list --domain cahut.example.com
sudo -u cahut /usr/local/bin/cahut token revoke --domain cahut.example.com --id 3
```

[Personal tokens](/connect/tokens/) says what a token allows and how to keep it.

## Read the Logbook

```bash
sudo -u cahut /usr/local/bin/cahut logbook --domain cahut.example.com --limit 50
sudo -u cahut /usr/local/bin/cahut logbook --server
```

The first command prints the latest entries of a Cahut's Logbook, the second those of the server. See [The Logbook](/concepts/knock-first-and-logbook/#the-logbook).

## Print the version

```bash
/usr/local/bin/cahut version
```

```text
cahut 0.1.0 (commit 0123abc, 2026-10-21T08:00:00Z, go1.27.3 linux/amd64)
```

## When a command fails

A command that fails says why on its error output, and exits with code 1; a wrong flag or argument exits with code 2. When the input is at fault, the message says what to change:

```text
cahut: invalid input: handle "Ana": use 2 to 32 lowercase letters, digits and hyphens, starting with a letter and ending with a letter or a digit
```
