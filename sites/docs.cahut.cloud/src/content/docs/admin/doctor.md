---
title: cahut doctor
description: cahut doctor diagnoses a server of Cahut, from the clock to the certificate of each Cahut, and says what to fix.
---

`cahut doctor` checks a server of Cahut and says, line by line, what it found and what to do. `install.sh` runs it at the end of each installation. Run it as root, so that it checks the firewall and SELinux too:

```bash
sudo /usr/local/bin/cahut doctor
```

Each line starts with `ok`, `warn` (it works, but look at it), `fail` (it does not work) or `info`. The command exits with 1 when a check fails.

## What it checks

| Topic | What it looks at |
| --- | --- |
| System | The system and its version: Cahut is tested on AlmaLinux 9 and 10, and Ubuntu 22.04 and 24.04 |
| Clock | Whether the clock is synchronized: certificates need the right time |
| Service | Whether the `cahut` service runs and starts at boot, and the version that answers, against the binary on disk |
| Data | The owner of `/var/lib/cahut`, and its free space |
| Ports | Whether the service listens on ports 80 and 443 |
| Firewall | Whether firewalld or ufw, when one is active, lets in ports 80 and 443 |
| SELinux | Its mode, and the label of the binary, which systemd needs |
| Proxy | When the service reads the [PROXY protocol](/admin/service/#behind-a-proxy): the proxies it trusts, and whether the last connection of one came without a valid header |
| Each Cahut | Whether it opens with its migrations, what its domain resolves to, the certificate it serves and its expiry, and whether `https://<domain>/` answers from the server |
| Let's Encrypt, GitHub | Whether the server reaches them |
| Updates | The daily update, the last update and a newer release. See [Updates](/admin/updates/) |
| Hosting provider | A reminder: ports 80 and 443 must be open in its firewall too |

`cahut doctor` does not open the databases: it asks the running service, on the server itself, at `http://127.0.0.1/-/healthz`, which answers there only with the version and the state of each Cahut.

## What it says, and what to do

| Line | What to do |
| --- | --- |
| `fail Clock Not synchronized` | `sudo timedatectl set-ntp true` |
| `fail Service The cahut service is failed` | `journalctl -u cahut` says why it stopped; `sudo systemctl start cahut` |
| `fail Ports Nothing listens on port 443` | Another program may hold the port: `sudo ss -ltnp` |
| `fail Firewall firewalld does not let in https` | `sudo firewall-cmd --permanent --add-service=http --add-service=https && sudo firewall-cmd --reload` |
| `fail SELinux … not bin_t` | `sudo restorecon /usr/local/bin/cahut` |
| `fail Proxy HTTPS on 127.0.0.1:8443 has refused every connection of 127.0.0.1` | The proxy sends no header of the PROXY protocol, or not a valid one: turn it on in the proxy, with `proxy_protocol on` in nginx, or empty `CAHUT_HTTPS_PROXY_PROTOCOL` and restart the service. The line says why the last connection was refused |
| `warn Proxy HTTPS on 127.0.0.1:8443 refused the last connection of 127.0.0.1` | A connection from the address of the proxy came without a header: another program of that address connects to Cahut directly, or the proxy sends the header on some connections only |
| `fail cahut.example.com does not resolve` | Add an A record, or AAAA, that points to the server |
| `warn cahut.example.com Points to 198.51.100.1` | Fine behind NAT; otherwise, the DNS record points elsewhere |
| `fail cahut.example.com No certificate yet` | Let's Encrypt must reach port 80 of the domain: check the DNS record and the firewall of the hosting provider |
| `warn … not one that browsers trust` | The certificate comes from the staging environment of Let's Encrypt: run `install.sh` again without `--staging` |
| `warn … does not answer from this machine` | Check the DNS record and the firewall of the hosting provider; some providers do not let a server reach its own public address |
| `warn Updates … was rolled back` | The new version failed its check; the reason follows. `journalctl -u cahut-update` says more |
