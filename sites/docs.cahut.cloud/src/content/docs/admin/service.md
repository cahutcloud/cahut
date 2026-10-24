---
title: The cahut service
description: Where Cahut keeps its files, how the systemd service runs, its settings, logs, certificates and health check, and how to back it up.
---

`install.sh` installs Cahut as the `cahut` systemd service. The service runs `cahut serve`, which serves every Cahut of the server: their MCP endpoints and their Windows.

## Files

| Path | What it holds |
| --- | --- |
| `/usr/local/bin/cahut` | The binary |
| `/etc/cahut/cahut.env` | The settings of the service |
| `/etc/cahut/update.env` | The token that reads the releases of a private fork, when the installer got one, for root only |
| `/etc/systemd/system/cahut.service` | The systemd unit |
| `/etc/systemd/system/cahut-update.*` | The updater: its service, its daily timer, and the path unit that starts it when the console asks |
| `/var/lib/cahut/` | The data: the records of the server, every Cahut with its Windows, and the certificates |
| `/var/lib/cahut-update/` | The state of updates, the previous binary and the copy of the databases before the last update |

[Cahuts and servers](/concepts/cahuts/#on-disk) shows what the data folder holds.

## Settings

`/etc/cahut/cahut.env` sets the environment of the service:

```sh
CAHUT_DATA=/var/lib/cahut
CAHUT_ACME_EMAIL=you@example.com
CAHUT_ACME_CA=
# What the daily update does: patch, all, check or off.
CAHUT_UPDATES=patch
```

| Variable | What it sets |
| --- | --- |
| `CAHUT_DATA` | The data folder |
| `CAHUT_ACME_EMAIL` | The contact of the Let's Encrypt account, optional |
| `CAHUT_ACME_CA` | Empty for Let's Encrypt, `staging` for its staging environment, or the address of another ACME directory |
| `CAHUT_UPDATES` | What the daily update does. See [Updates](/admin/updates/) |

After a change to the file, restart the service: `sudo systemctl restart cahut`. `install.sh` writes the file again when you pass it `--email`, `--staging` or `--updates`, and keeps the values you do not give again.

## The service

```bash
sudo systemctl status cahut
sudo systemctl restart cahut
```

The service runs as the `cahut` user and restarts 5 seconds after a failure. It listens on port 80, which answers the challenges of Let's Encrypt and redirects to HTTPS, and on port 443, for the Windows, MCP, OAuth, and the console and its [API](/reference/http-api/) under `/console/`. When it stops, it lets the requests in progress finish, for up to 10 seconds.

The unit limits what the service can do: it keeps one capability, to listen on ports below 1024, cannot gain others, sees the system as read-only except its data folder, and has no access to home folders, devices or kernel settings.

## Logs

The service writes to the journal:

```bash
journalctl -u cahut -f
```

At start, it names the Cahuts it serves. It logs the errors that it hides from the AIs, with the Cahut, the tool and the person, and its certificates, on lines named `certificates`. The actions of AIs and people are in the [Logbook](/concepts/knock-first-and-logbook/#the-logbook), not in the journal.

## Certificates

The service gets a certificate from Let's Encrypt for the domain of each Cahut, and renews it. A Cahut at the IP address of the server gets a certificate for that address, which lasts six days and is renewed every few days. It obtains those of the Cahuts it knows when it starts, and that of a Cahut declared while it runs at the first HTTPS connection to its domain. When Let's Encrypt fails, for instance because the DNS record does not point to the server yet, the service tries again in the background.

It asks for no certificate for a name that no Cahut declares. The certificates and the keys of the ACME account are in `/var/lib/cahut/certificates`.

## Health check

Every Cahut answers `ok` at `/-/healthz` while the service runs:

```bash
curl https://cahut.example.com/-/healthz
```

On the server itself, port 80 answers at a loopback address with the version and the state of every Cahut, which it opens with their migrations. The updater checks a new version so, and [cahut doctor](/admin/doctor/) reads it:

```bash
curl http://127.0.0.1/-/healthz
```

```json
{"version":"0.5.0","ok":true,"cahuts":[{"domain":"cahut.example.com","ok":true}]}
```

## Back up

Cahut has no backups of its own yet: they come later, to S3 or SFTP storage. Until then, take snapshots of the server at your provider, or copy the data folder while the service is stopped:

```bash
sudo systemctl stop cahut
sudo tar -C /var/lib -czf cahut-backup.tar.gz cahut
sudo systemctl start cahut
```

Stop the service first: a copy of a database in use may not open.

## Update

The server updates itself every day, with a rollback when a new version fails: see [Updates](/admin/updates/).
