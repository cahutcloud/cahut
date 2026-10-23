---
title: Install on a server
description: Install Cahut as a systemd service on a Linux server with install.sh, and declare your first Cahut.
---

`install.sh` installs Cahut on a Linux server as the `cahut` systemd service, declares a Cahut at your domain, and starts the service. Run it again to update Cahut or to add a Cahut.

## Before you start

You need:

- **A Linux server with systemd**, on amd64 or arm64, where you can run commands as root. Cahut targets AlmaLinux and Ubuntu, from the provider's standard image. Plan 2 vCPU, 4 GB of RAM and 40 GB of disk, the size the project aims at for its launch: today's Cahut needs less, and the disk fills with the Windows you publish.
- **A domain or a subdomain for each Cahut**, such as `cahut.example.com`, with an A or AAAA record that points to the server. No wildcard record is needed.
- **Ports 80 and 443 open** to the internet, in the provider's firewall too. Port 80 answers the challenges of Let's Encrypt and redirects to HTTPS. Port 443 serves everything else: the Windows, MCP, OAuth, and the console under `/console/`.
- **curl**, to download the release.

Create the DNS record first: the service asks Let's Encrypt for the certificate of each domain as soon as it starts.

## Run the installer

Download the script, then run it as root with the domain of your first Cahut:

```bash
curl -fsSLO https://raw.githubusercontent.com/getcahut/cahut/main/install.sh
sudo sh install.sh --domain cahut.example.com --email you@example.com
```

The email address is optional: Let's Encrypt writes to it about your certificates. The script says what it does, then prints the next steps:

```text
cahut: downloading Cahut 0.2.0 for linux/amd64
cahut: creating the cahut system user
cahut: writing /etc/cahut/cahut.env
cahut: declaring the Cahut of cahut.example.com
cahut: opening ports 80 and 443 in firewalld
cahut: cahut 0.2.0 (commit 39dd726, 2026-10-23T06:08:00Z, go1.27.3 linux/amd64) is running
  https://cahut.example.com/

Next steps for the Cahut of cahut.example.com:

  1. Create your account, as an administrator of the Cahut. The command
     prints a link where you choose your password:
       sudo -u cahut /usr/local/bin/cahut user create --domain cahut.example.com --handle you --name "Your Name" --admin
  2. Open that link: it signs you in to the console, at https://cahut.example.com/console/
  3. In the console, AIs and access creates a token for your AI, with the
     command that adds the Cahut to Claude Code.

To administer the server itself, create an administrator of the server:
       sudo -u cahut /usr/local/bin/cahut admin create --handle you --name "Your Name"

Ports 80 and 443 must be open in the firewall of your hosting provider
too, if it has one.

The service logs go to the journal: journalctl -u cahut
```

[Connect Claude Code](/connect/claude-code/) explains these three steps, and [The console](/admin/console/) what you find there.

## What the installer does

1. It downloads the release for the server's architecture from GitHub, the latest one unless you name another, and checks its SHA-256 checksum. With `--binary`, it takes the binary you give instead.
2. It checks that the binary runs on the server.
3. It creates the `cahut` system user, with no shell, and the data folder `/var/lib/cahut`, which only that user and its group can read.
4. It writes the settings file `/etc/cahut/cahut.env`, unless the file exists already.
5. It writes the systemd unit `/etc/systemd/system/cahut.service` and installs the binary as `/usr/local/bin/cahut`. On a server with SELinux, it restores the labels of both files.
6. It declares the Cahut of `--domain`, unless the server holds it already.
7. It opens ports 80 and 443 in firewalld or ufw, when one of them is active. Earlier versions opened port 8443 for the console too: the script leaves that port as it is and prints the command that closes it.
8. It enables and restarts the service, then checks that it runs.

The service runs as the `cahut` user, with the one capability it needs to listen on ports 80 and 443, and can write only to its data folder. See [The cahut service](/admin/service/).

## Options

| Option | What it does |
| --- | --- |
| `--domain DOMAIN` | Declares the Cahut of `DOMAIN`, such as `cahut.example.com`. It must point to this server. The first run needs it. |
| `--email EMAIL` | Sets the contact of the Let's Encrypt account. Optional. |
| `--staging` | Uses the staging environment of Let's Encrypt, for tests. Browsers do not trust its certificates. |
| `--version TAG` | Installs this release, such as `v0.2.0`, instead of the latest. |
| `--binary PATH` | Installs this `cahut` binary instead of downloading a release. |
| `--help` | Prints the options. |

Two environment variables choose where releases come from: `CAHUT_REPO`, the GitHub repository of the releases, `getcahut/cahut` by default, and `GITHUB_TOKEN`, a token that can read the releases of a private repository. Give them on the command line of `sudo`:

```bash
sudo CAHUT_REPO=you/cahut sh install.sh --domain cahut.example.com
```

## Add a Cahut

Point the DNS record of the new domain to the server, then run the script again with that domain:

```bash
sudo sh install.sh --domain docs.example.org
```

The script also installs the latest release, as for an update. To declare the domain without updating, use the command line instead: the running service gets the new certificate at the first HTTPS connection to that domain.

```bash
sudo -u cahut /usr/local/bin/cahut domain add docs.example.org
```

## Update Cahut

Run the script again. It installs the latest release, or the one of `--version`, and restarts the service. The data and the Cahuts stay as they are.

```bash
sudo sh install.sh
```

The settings file stays as it is too, unless you pass `--email` or `--staging`: the script then writes it again from the options you give. A run with `--email` and without `--staging` therefore returns to the production environment of Let's Encrypt.

To go back to an earlier release, run the script with its `--version`.

## Check that it runs

```bash
systemctl status cahut
curl https://cahut.example.com/-/healthz
```

The health check answers `ok`. The address of the Cahut answers with one line:

```text
This is cahut.example.com, a Cahut. AIs connect to it over MCP at /mcp.
```

If the service does not start, its logs say why: `journalctl -u cahut`.

## Remove Cahut

There is no uninstaller yet. To remove the service and keep the data:

```bash
sudo systemctl disable --now cahut
sudo rm /etc/systemd/system/cahut.service /usr/local/bin/cahut
sudo rm -r /etc/cahut
sudo systemctl daemon-reload
```

`/var/lib/cahut` then still holds every Cahut, its Windows and its Logbook, and the `cahut` user still exists. Deleting that folder deletes them for good.
