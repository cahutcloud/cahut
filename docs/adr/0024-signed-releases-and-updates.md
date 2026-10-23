# 24. Releases are signed, and servers update themselves with a rollback

- Status: Accepted
- Date: 2026-10-23

## Context

[ADR 7](0007-module-path-repository-and-releases.md) builds releases with GoReleaser when a `v*` tag is pushed, as drafts that a person publishes, with a SHA-256 checksum file. It leaves their signature to batch 7, for the updater. Today `install.sh` checks the checksum, which proves that the archive arrived whole, not that the project made it: whoever can change a release can change its checksum file too.

A server updates when its administrator runs `install.sh` again. The script replaces the binary and restarts the service, without checking that the new version starts, and without a way back. A release whose migration fails on one Cahut, or that no longer binds its ports, leaves the server down until someone notices.

The `cahut` service runs as an unprivileged user, with `ProtectSystem=strict` and `NoNewPrivileges` ([install.sh](../../install.sh)). It cannot replace its own binary or restart itself, and it should not gain that power: a flaw in the code that serves Windows written by AIs would then reach the whole machine.

## Decision

- **An Ed25519 key signs the checksum file of each release**, as `checksums.txt.sig`, the raw 64 bytes of the signature. The checksum file names every archive, so one signature covers them all. The private key is a secret of the repository, `CAHUT_RELEASE_KEY`, read by the release workflow only; the lead maintainer keeps an offline copy. The public key is embedded in the binary and written in `install.sh`. A new key ships in a release signed by the old one, and binaries accept every key of their list. Releases are signed from 0.4.0 on.
- **`install.sh` checks the signature with `openssl`**, then the checksum. OpenSSL 3.0 or later verifies Ed25519 signatures, and AlmaLinux 9 and 10, Ubuntu 22.04 and 24.04 ship it. Without `openssl`, the script stops and says how to install it. `cahut update` checks the same way, with the standard library.
- **`cahut update`, run as root, updates the server**:
  1. It reads the published releases of the repository, and takes the latest one, or the one of `--version`. It never installs a version older than the one running.
  2. It downloads the archive for the machine, the checksum file and its signature; checks the signature, then the checksum; extracts the binary and runs `cahut version` to see that it starts on this machine.
  3. It stops the service, copies the databases (`server.db` and the database of each Cahut) to a snapshot, keeps the running binary, installs the new one, and starts the service.
  4. It waits up to 60 seconds for the service to answer on the loopback address with the new version, having opened every Cahut, which runs its migrations.
  5. When the service does not answer so, it stops it, puts back the previous binary and the snapshot of the databases, starts it again and checks it the same way, then reports the update as rolled back, with the reason.
- **`cahut update --binary <file>`** installs a local binary with the same snapshot, check and rollback, without a signature, as `install.sh --binary` does: the administrator vouches for it.
- **Updates are automatic within limits.** A systemd timer runs `cahut update --unattended` every day, at a random time. `CAHUT_UPDATES`, in `/etc/cahut/cahut.env`, says what it does: `patch`, the default, installs on its own the patch releases of the minor version running (0.4.0 to 0.4.3) and only reports the others; `all` installs any newer release; `check` only reports; `off` asks GitHub nothing.
- **Root updates; the service only reads and asks.** The state of updates lives in `/var/lib/cahut-update/`, owned by root and readable by the `cahut` group: `status.json`, with the last check and the last update, the previous binary and the snapshot of the databases. The service never calls GitHub. From the console, an administrator of the server asks for a check or an update: the service writes `update-request` in its data directory, with a version number or `check`, and a systemd `.path` unit starts the updater. The updater reads that file without following links, 64 bytes at most, accepts a version newer than the one running or `check`, and nothing else. A compromised service can thus ask for a signed release newer than its own, and no more.
- **Root never opens the databases.** SQLite would create root-owned journal files that the service could not open. The updater copies files while the service is stopped, and gives back their owner; `cahut doctor` asks the running service instead.
- **A database newer than the binary does not open.** `store.Open` refuses a database whose schema version is above the last migration of the binary, and says that a newer version of Cahut made it. A binary put back by hand, days after an update, stops there instead of running on a schema it does not know.
- **The Logbook of the server records updates.** At start, the service writes each update that `status.json` reports and that the Logbook does not have yet: from which version to which, by whom (the timer, the console or the command line), and how it ended. The request made in the console is logged as an operation, `request_update`.
- **While the repository is private**, `install.sh` keeps the `GITHUB_TOKEN` it was given in `/etc/cahut/update.env`, readable by root only, for the updater. The file goes away when the repository opens.

## Consequences

- Changing a release, or the checksum file, no longer goes unnoticed: a server checks what it installs against a key that the repository does not serve.
- Losing the private key means that servers can no longer verify new releases: they need a release signed with the old key that names the new one, or a new install. Its offline copy is kept with the other keys of the project, such as those of its domain and of its GitHub organization.
- A failed update leaves the server on its previous version within about a minute, with the reason in `status.json`, in `cahut doctor` and in the console.
- The snapshot of the databases restores what the new version may have migrated. Window files are not copied: migrations do not touch them. What happened during the minute of the failed version is lost.
- A server stays on its minor version until its administrator moves it on, unless `CAHUT_UPDATES=all`. Before 1.0, every batch brings a minor version, so most updates wait for a person.
- The updater runs as root, outside the hardening of the service. Its input from the service is a version number at most, and what it installs is signed.

## Alternatives considered

- **Sigstore, keyless signatures from the release workflow.** No key to keep, but verifying a Sigstore bundle in Go takes a large set of dependencies and trust roots fetched over the network, and `install.sh` would need `cosign` on the server.
- **minisign or signify.** Small and well known, but neither ships with AlmaLinux or Ubuntu: `install.sh` would need to download a verifier before it could verify anything.
- **GitHub artifact attestations.** Checking them needs the `gh` command and an account on the server.
- **The service updating itself.** Simpler, but the service would need to write its own binary and restart itself, which is the power its hardening takes away.
- **Packages from a dnf and an apt repository.** Signed by the systems' own tools and updated with the rest of the machine, at the cost of running two repositories, and without the health check and the rollback, which the V0 wants. Possible after the V0.
- **Every update automatic.** One less step for administrators, but a server holds the work of its people: a new minor version may change what AIs see, and should be a choice before 1.0.
