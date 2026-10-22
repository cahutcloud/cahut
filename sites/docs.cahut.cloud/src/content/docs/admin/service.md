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
| `/etc/systemd/system/cahut.service` | The systemd unit |
| `/var/lib/cahut/` | The data: the records of the server, every Cahut with its Windows, and the certificates |

[Cahuts and servers](/concepts/cahuts/#on-disk) shows what the data folder holds.

## Settings

`/etc/cahut/cahut.env` sets the environment of the service:

```sh
CAHUT_DATA=/var/lib/cahut
CAHUT_ACME_EMAIL=you@example.com
CAHUT_ACME_CA=
```

| Variable | What it sets |
| --- | --- |
| `CAHUT_DATA` | The data folder |
| `CAHUT_ACME_EMAIL` | The contact of the Let's Encrypt account, optional |
| `CAHUT_ACME_CA` | Empty for Let's Encrypt, `staging` for its staging environment, or the address of another ACME directory |

After a change to the file, restart the service: `sudo systemctl restart cahut`. `install.sh` writes the file again when you pass it `--email` or `--staging`.

## The service

```bash
sudo systemctl status cahut
sudo systemctl restart cahut
```

The service runs as the `cahut` user and restarts 5 seconds after a failure. It listens on port 80, which answers the challenges of Let's Encrypt and redirects to HTTPS, on port 443, for the Windows and MCP, and on port 8443, for the console and its [API](/reference/http-api/), with the same certificates. When it stops, it lets the requests in progress finish, for up to 10 seconds.

The unit limits what the service can do: it keeps one capability, to listen on ports below 1024, cannot gain others, sees the system as read-only except its data folder, and has no access to home folders, devices or kernel settings.

## Logs

The service writes to the journal:

```bash
journalctl -u cahut -f
```

At start, it names the Cahuts it serves. It logs the errors that it hides from the AIs, with the Cahut, the tool and the person, and its certificates, on lines named `certificates`. The actions of AIs and people are in the [Logbook](/concepts/knock-first-and-logbook/#the-logbook), not in the journal.

## Certificates

The service gets a certificate from Let's Encrypt for the domain of each Cahut, and renews it. It obtains those of the Cahuts it knows when it starts, and that of a Cahut declared while it runs at the first HTTPS connection to its domain. When Let's Encrypt fails, for instance because the DNS record does not point to the server yet, the service tries again in the background.

It asks for no certificate for a name that no Cahut declares. The certificates and the keys of the ACME account are in `/var/lib/cahut/certificates`.

## Health check

Every Cahut answers `ok` at `/-/healthz` while the service runs:

```bash
curl https://cahut.example.com/-/healthz
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

Run `install.sh` again: see [Update Cahut](/install/server/#update-cahut).
