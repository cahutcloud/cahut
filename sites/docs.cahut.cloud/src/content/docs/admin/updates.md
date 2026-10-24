---
title: Updates
description: How a server of Cahut updates itself from its signed releases, with a health check and an automatic rollback, and how to update it by hand or from the console.
---

A server of Cahut installs its new releases from GitHub. Each release is signed: the updater checks the signature, copies the databases, installs the new binary, checks that the service answers with it, and goes back to the previous version on its own when it does not.

## Every day

A timer of systemd, `cahut-update.timer`, runs the updater once a day, at a random time between 3 and 6 in the morning. `CAHUT_UPDATES`, in `/etc/cahut/cahut.env`, says what it does:

| Value | What the daily run does |
| --- | --- |
| `patch` | The default. Installs the patch releases of the minor version running, such as 0.4.3 after 0.4.0, and reports the others |
| `all` | Installs any newer release |
| `check` | Reports newer releases, installs nothing |
| `off` | Asks GitHub nothing |

A new minor version may change what AIs see: before 1.0, it waits for a person, unless you choose `all`. Set the value with `install.sh --updates`, or edit the file.

## Update by hand

```bash
sudo /usr/local/bin/cahut update --check
sudo /usr/local/bin/cahut update
sudo /usr/local/bin/cahut update --version 0.4.2
```

`--check` says whether a newer release is out. Without it, the updater installs the newest release, or the one of `--version`. It never installs a version older than the one running.

`--binary` installs a binary you give, such as a build of your own, with the same copy of the databases, health check and rollback, and without a signature: you vouch for it.

## In the console

The administrators of the server find **This server › Updates**: the version running, the newest release and a link to its notes, the last check, the last update and how it ended. *Check now* and *Update to 0.4.2* ask the updater, which runs within a minute. The console may stop answering for a few seconds while the service restarts.

## What an update does

1. It reads the published releases of the repository. Drafts, which the project has not published, stay out.
2. It downloads the archive for the server, the checksum file and its signature, checks the signature against the keys of Cahut embedded in the binary, then the checksum, and runs the new binary once.
3. It stops the service, copies `server.db` and the database of each Cahut to `/var/lib/cahut-update/snapshot`, keeps the binary running as `/var/lib/cahut-update/previous/cahut`, installs the new one, and starts the service.
4. It waits up to a minute for the service to answer on `http://127.0.0.1/-/healthz` with the new version, every Cahut open, which runs its migrations.
5. If the service does not answer so, it stops it, puts back the previous binary and the copy of the databases, starts it again, and records the update as rolled back, with the reason.

What happened during that minute is lost. The files of Windows are not copied: migrations do not touch them. The copy of the databases is not a backup: it holds the state before the last update only.

## Where you see what happened

- The **Logbook of the server** records each update, from which version to which, who started it (the timer, the console or the command line) and how it ended.
- [cahut doctor](/admin/doctor/) says when the last update was rolled back or failed, and when a newer release is out.
- `/var/lib/cahut-update/status.json` holds the last check and the last updates; `journalctl -u cahut-update` holds what the updater printed.

## Signatures

An Ed25519 key signs the checksum file of each release, `checksums.txt.sig`. The checksum file names every archive, so one signature covers them all. The public keys are in the binary and in `install.sh`; the private key stays with the project. To check a release by hand with OpenSSL 3, with the key of `install.sh` saved as `release-key.pem`:

```bash
openssl pkeyutl -verify -pubin -inkey release-key.pem -rawin -in checksums.txt -sigfile checksums.txt.sig
sha256sum --ignore-missing -c checksums.txt
```

## Going back by hand

The updater goes back on its own when a new version fails at once. To return to an earlier version later, reinstall it with `install.sh --version`: a database that a newer version changed does not open with an older binary, which says so, so restore a snapshot of the server from before the update.

## Rights

The service runs as the unprivileged `cahut` user: it can neither replace its binary nor restart itself, and it never calls GitHub. The updater runs as root, in `cahut-update.service`. The console asks it through a file of the data folder, `update-request`, which a `.path` unit of systemd watches; the updater reads it without following links and accepts only a check or a version newer than the one running. When the releases come from a private fork, the token that reads them is in `/etc/cahut/update.env`, readable by root only.
