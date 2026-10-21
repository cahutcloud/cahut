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

## Create a person

```bash
sudo -u cahut /usr/local/bin/cahut user create --domain cahut.example.com --handle ana --name "Ana Lima"
```

```text
Created ana (Ana Lima), member of cahut.example.com, with a personal Room at cahut.example.com/ana.
```

The handle names the person and their personal Room: 2 to 32 lowercase letters, digits and hyphens, starting with a letter and ending with a letter or a digit. It must be free, as the name of a person and of a Room, and cannot be `mcp`, `api`, `admin`, `console` or `help`. `--admin` makes the person an administrator of the Cahut.

There is no command yet to list, change or remove people.

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
