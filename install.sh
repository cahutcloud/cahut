#!/bin/sh
# Install or update Cahut on a Linux server with systemd.
#
#   sudo sh install.sh --domain cahut.example.com [--email you@example.com]
#   sudo sh install.sh --ip 203.0.113.7
#
# The script downloads a release from GitHub and checks its signature and
# its checksum, or takes a binary you give it with --binary. It then creates the cahut system
# user, the data directory /var/lib/cahut, the settings file
# /etc/cahut/cahut.env and the cahut service, declares the Cahut of
# --domain, opens ports 80 and 443 in firewalld or ufw when one is active,
# starts the service, and runs cahut doctor to say what is left to fix. A
# timer of systemd then installs new releases every day, as --updates
# says, with a rollback if one fails.
#
# Options:
#   --domain DOMAIN   domain of a Cahut to declare on this server, such as
#                     cahut.example.com; it must point to this server. Run
#                     the script again with another domain to add a Cahut.
#   --ip ADDRESS      without a domain, the public IPv4 address of this
#                     server, where a Cahut answers with a certificate of
#                     its own; it moves to a domain later. Without --domain
#                     and --ip, the script takes the address of the default
#                     route when it is public
#   --email EMAIL     contact of the Let's Encrypt account (optional)
#   --staging         use Let's Encrypt's staging environment, for tests
#   --version TAG     release to install, such as v0.4.0 (default: latest);
#                     releases are signed from v0.4.0 on
#   --binary PATH     install this cahut binary instead of downloading one
#   --updates POLICY  what the daily update does: patch (the default)
#                     installs the patch releases of the minor version
#                     running, all installs any newer release, check only
#                     reports them, off asks GitHub nothing
#
# Environment:
#   CAHUT_REPO        GitHub repository of the releases (default: getcahut/cahut)
#   GITHUB_TOKEN      token to read the releases of a private repository,
#                     kept for the updates in /etc/cahut/update.env
#
# Run it again to change its settings. To update Cahut, run
# sudo /usr/local/bin/cahut update.

set -eu

repo="${CAHUT_REPO:-getcahut/cahut}"
domain=""
ip=""
email=""
ca=""
tag="latest"
binary=""
updates=""

bin_path=/usr/local/bin/cahut
data_dir=/var/lib/cahut
conf_dir=/etc/cahut
env_file=$conf_dir/cahut.env
unit_file=/etc/systemd/system/cahut.service
update_env=$conf_dir/update.env
state_dir=/var/lib/cahut-update
units=/etc/systemd/system

say() { printf 'cahut: %s\n' "$*"; }
die() { printf 'cahut: %s\n' "$*" >&2; exit 1; }

while [ $# -gt 0 ]; do
	case "$1" in
	--domain) domain="${2:?--domain needs a value}"; shift 2 ;;
	--ip) ip="${2:?--ip needs a value}"; shift 2 ;;
	--email) email="${2:?--email needs a value}"; shift 2 ;;
	--staging) ca=staging; shift ;;
	--version) tag="${2:?--version needs a value}"; shift 2 ;;
	--binary) binary="${2:?--binary needs a value}"; shift 2 ;;
	--updates) updates="${2:?--updates needs a value}"; shift 2 ;;
	-h | --help) sed -n '2,/^$/s/^# \{0,1\}//p' "$0"; exit 0 ;;
	*) die "unknown option $1 (see --help)" ;;
	esac
done

[ "$(id -u)" -eq 0 ] || die "run this script as root, for instance with sudo"
command -v systemctl >/dev/null || die "systemd is required"
# public_ip ADDRESS succeeds when ADDRESS is an IPv4 address that the world
# can reach: not private, loopback, link-local or shared. cahut checks it
# again when it declares it.
public_ip() {
	printf '%s' "$1" | grep -Eq '^[0-9]{1,3}(\.[0-9]{1,3}){3}$' || return 1
	case "$1" in
	10.* | 127.* | 169.254.* | 192.168.* | 0.*) return 1 ;;
	172.1[6-9].* | 172.2[0-9].* | 172.3[01].*) return 1 ;;
	100.6[4-9].* | 100.[7-9][0-9].* | 100.1[01][0-9].* | 100.12[0-7].*) return 1 ;;
	esac
}

[ -z "$domain" ] || [ -z "$ip" ] || die "give --domain or --ip, not both"
if [ -n "$ip" ]; then
	public_ip "$ip" || die "$ip is not a public IPv4 address"
	domain=$ip
elif [ -z "$domain" ] && [ ! -f "$env_file" ]; then
	# A first install without a domain: the Cahut takes the address of the
	# server. Behind NAT, the machine does not know its public address.
	ip=$(ip -4 route get 1.1.1.1 2>/dev/null | sed -n 's/.* src \([0-9.]*\).*/\1/p')
	public_ip "$ip" ||
		die "give the domain of your first Cahut with --domain, or the public IPv4 address of this server with --ip"
	say "no --domain: the Cahut takes the address of this server, $ip"
	domain=$ip
fi
case "$updates" in
"" | patch | all | check | off) ;;
*) die "--updates takes patch, all, check or off" ;;
esac
if [ -n "$domain" ] && [ -z "$ip" ]; then
	domain=$(printf '%s' "$domain" | tr 'A-Z' 'a-z')
	printf '%s' "$domain" | grep -Eq '^[a-z0-9]([a-z0-9-]*[a-z0-9])?(\.[a-z0-9]([a-z0-9-]*[a-z0-9])?)+$' ||
		die "$domain is not a domain name"
fi

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

# fetch URL FILE [ACCEPT] downloads URL to FILE, with GITHUB_TOKEN when set.
fetch() {
	if [ -n "${GITHUB_TOKEN:-}" ]; then
		curl -fsSL -H "Authorization: Bearer $GITHUB_TOKEN" -H "Accept: ${3:-application/json}" -o "$2" "$1"
	else
		curl -fsSL -H "Accept: ${3:-application/json}" -o "$2" "$1"
	fi
}

# The public keys that sign the releases: the checksum file of each release
# is signed with one of them. They are those of internal/release/release-keys.pem.
# See docs/adr/0024-signed-releases-and-updates.md.
release_keys() {
	cat <<'EOF'
-----BEGIN PUBLIC KEY-----
MCowBQYDK2VwAyEAEKEIUk1PktFCPrG//Y8BJ9CMQnNLTn9ibkurcJxQQi4=
-----END PUBLIC KEY-----
EOF
}

# verify_signature FILE SIGNATURE checks that one of the release keys signed
# FILE. OpenSSL 3 verifies Ed25519 signatures of raw data.
verify_signature() {
	release_keys | awk -v dir="$tmp" '
		/-----BEGIN PUBLIC KEY-----/ { n++ }
		n { print > (dir "/release-key-" n ".pem") }
	'
	for key in "$tmp"/release-key-*.pem; do
		if openssl pkeyutl -verify -pubin -inkey "$key" -rawin -in "$1" -sigfile "$2" >/dev/null 2>&1; then
			return 0
		fi
	done
	return 1
}

# asset_url NAME prints the API address of a release asset, from the
# release description in $tmp/release.txt. In each asset, the GitHub API
# gives the url member before the name.
asset_url() {
	awk -v name="$1" '
		/^ *"url": *"https:\/\/api\.github\.com\/repos\/[^"]*\/releases\/assets\/[0-9]+" *$/ {
			url = $0; sub(/^ *"url": *"/, "", url); sub(/" *$/, "", url)
		}
		/^ *"name": *"/ {
			n = $0; sub(/^ *"name": *"/, "", n); sub(/" *$/, "", n)
			if (n == name) { print url; exit }
		}
	' "$tmp/release.txt"
}

if [ -n "$binary" ]; then
	[ -f "$binary" ] || die "no file $binary"
	cp "$binary" "$tmp/cahut"
else
	command -v curl >/dev/null || die "curl is required to download Cahut"
	command -v openssl >/dev/null ||
		die "openssl is required to check the signature of the release: install it with dnf install openssl, or apt install openssl"
	case "$(uname -m)" in
	x86_64 | amd64) arch=amd64 ;;
	aarch64 | arm64) arch=arm64 ;;
	*) die "no release for $(uname -m)" ;;
	esac
	if [ "$tag" = latest ]; then
		release="https://api.github.com/repos/$repo/releases/latest"
	else
		release="https://api.github.com/repos/$repo/releases/tags/$tag"
	fi
	fetch "$release" "$tmp/release.json" ||
		die "cannot read the release from $repo (private repository? set GITHUB_TOKEN)"
	# One JSON member per line, whether the API indents its answer or not.
	tr '{},' '\n\n\n' <"$tmp/release.json" >"$tmp/release.txt"
	version=$(sed -n 's/^ *"tag_name": *"v\{0,1\}\([^"]*\)" *$/\1/p' "$tmp/release.txt" | head -n 1)
	[ -n "$version" ] || die "cannot read the version of the release"
	archive="cahut_${version}_linux_${arch}.tar.gz"
	say "downloading Cahut $version for linux/$arch"
	for file in "$archive" checksums.txt checksums.txt.sig; do
		url=$(asset_url "$file")
		if [ -z "$url" ] && [ "$file" = checksums.txt.sig ]; then
			die "release $version is not signed: install v0.4.0 or later, or give a binary with --binary"
		fi
		[ -n "$url" ] || die "the release has no $file"
		fetch "$url" "$tmp/$file" application/octet-stream || die "cannot download $file"
	done
	verify_signature "$tmp/checksums.txt" "$tmp/checksums.txt.sig" ||
		die "the signature of release $version does not match the keys of Cahut: it was not made by the project"
	(cd "$tmp" && grep " $archive\$" checksums.txt | sha256sum -c --quiet -) || die "checksum mismatch for $archive"
	say "signature and checksum checked"
	tar -xzf "$tmp/$archive" -C "$tmp" cahut
fi
chmod 0755 "$tmp/cahut"
"$tmp/cahut" version >/dev/null 2>&1 || die "the binary does not run on this machine"

if ! id cahut >/dev/null 2>&1; then
	say "creating the cahut system user"
	nologin=$(command -v nologin || echo /usr/sbin/nologin)
	useradd --system --user-group --home-dir "$data_dir" --no-create-home --shell "$nologin" cahut
fi
install -d -m 0750 -o cahut -g cahut "$data_dir"
install -d -m 0750 -o root -g cahut "$conf_dir"

# setting NAME prints the value of NAME in the settings file, if any.
setting() { sed -n "s/^$1=//p" "$env_file" 2>/dev/null | tail -n 1; }
[ -n "$email" ] || email=$(setting CAHUT_ACME_EMAIL)
[ -n "$ca" ] || ca=$(setting CAHUT_ACME_CA)
[ -n "$updates" ] || updates=$(setting CAHUT_UPDATES)
[ -n "$updates" ] || updates=patch
cat >"$tmp/cahut.env" <<EOF
# Settings of the cahut service. Run install.sh again to change them.
CAHUT_DATA=$data_dir
CAHUT_ACME_EMAIL=$email
CAHUT_ACME_CA=$ca
# What the daily update does: patch, all, check or off.
CAHUT_UPDATES=$updates
EOF
if [ "$(cat "$tmp/cahut.env")" != "$(cat "$env_file" 2>/dev/null)" ]; then
	say "writing $env_file"
	install -m 0640 -o root -g cahut "$tmp/cahut.env" "$env_file"
fi
# The updater runs as root, reads the releases with this token while the
# repository is private, and is the only one to read it.
if [ -n "${GITHUB_TOKEN:-}" ]; then
	(umask 077 && printf 'GITHUB_TOKEN=%s\n' "$GITHUB_TOKEN" >"$tmp/update.env")
	install -m 0600 -o root -g root "$tmp/update.env" "$update_env"
fi
# The state of updates: root writes it, the service reads it.
install -d -m 2750 -o root -g cahut "$state_dir"

cat >"$unit_file" <<EOF
[Unit]
Description=Cahut
Documentation=https://cahut.cloud
After=network-online.target
Wants=network-online.target

[Service]
User=cahut
Group=cahut
EnvironmentFile=$env_file
ExecStart=$bin_path serve
Restart=on-failure
RestartSec=5s
AmbientCapabilities=CAP_NET_BIND_SERVICE
CapabilityBoundingSet=CAP_NET_BIND_SERVICE
NoNewPrivileges=yes
ProtectSystem=strict
ReadWritePaths=$data_dir
ProtectHome=yes
PrivateTmp=yes
PrivateDevices=yes
ProtectKernelTunables=yes
ProtectKernelModules=yes
ProtectControlGroups=yes
RestrictAddressFamilies=AF_INET AF_INET6 AF_UNIX
RestrictNamespaces=yes
LockPersonality=yes
MemoryDenyWriteExecute=yes
SystemCallArchitectures=native
UMask=0027
# Past MemoryHigh, the service is slowed to give way to the rest of the
# machine, such as SSH; past MemoryMax, the kernel ends it, and systemd
# starts it again.
MemoryHigh=75%
MemoryMax=85%

[Install]
WantedBy=multi-user.target
EOF

# The updater: a daily timer, and a path unit that starts it when the
# console asks. See docs/adr/0024-signed-releases-and-updates.md.
cat >"$units/cahut-update.service" <<EOF
[Unit]
Description=Update Cahut from its signed releases
Documentation=https://docs.cahut.cloud/admin/updates/
After=network-online.target
Wants=network-online.target

[Service]
Type=oneshot
ExecStart=$bin_path update --unattended
EOF
cat >"$units/cahut-update.timer" <<EOF
[Unit]
Description=Look for a new release of Cahut every day

[Timer]
OnCalendar=*-*-* 03:00:00
RandomizedDelaySec=3h
Persistent=true

[Install]
WantedBy=timers.target
EOF
cat >"$units/cahut-update.path" <<EOF
[Unit]
Description=Update Cahut when its console asks

[Path]
PathExists=$data_dir/update-request

[Install]
WantedBy=paths.target
EOF

install -m 0755 "$tmp/cahut" "$bin_path"
if command -v restorecon >/dev/null; then
	# SELinux labels of the new files
	restorecon "$bin_path" "$unit_file" "$units"/cahut-update.* "$state_dir" || true
fi

# as_cahut runs a cahut command as the service user, on its data.
as_cahut() { runuser -u cahut -- "$bin_path" "$@" --data "$data_dir"; }

if [ -n "$domain" ] && ! as_cahut domain list | awk 'NR > 1 { print $2 }' | grep -qx "$domain"; then
	say "declaring the Cahut of $domain"
	as_cahut domain add "$domain" >/dev/null
fi

# The console is served under /console/ on port 443: port 8443, which
# releases before it opened, serves nothing. The script leaves it as it is
# and says how to close it.
if systemctl is-active --quiet firewalld 2>/dev/null; then
	say "opening ports 80 and 443 in firewalld"
	firewall-cmd --quiet --permanent --add-service=http --add-service=https
	firewall-cmd --quiet --reload
	if firewall-cmd --quiet --permanent --query-port=8443/tcp; then
		say "port 8443 is open but no longer used: firewall-cmd --permanent --remove-port=8443/tcp && firewall-cmd --reload"
	fi
elif command -v ufw >/dev/null && ufw status 2>/dev/null | grep -q '^Status: active'; then
	say "opening ports 80 and 443 in ufw"
	ufw allow 80/tcp >/dev/null
	ufw allow 443/tcp >/dev/null
	if ufw status | grep -q '^8443/tcp'; then
		say "port 8443 is open but no longer used: ufw delete allow 8443/tcp"
	fi
fi

systemctl daemon-reload
systemctl enable --quiet cahut
systemctl enable --quiet --now cahut-update.timer cahut-update.path
systemctl restart cahut
sleep 2
systemctl is-active --quiet cahut || die "the service did not start: see journalctl -u cahut"

say "$("$bin_path" version) is running"

# A new Cahut gets its certificate within seconds once Let's Encrypt
# reaches it: wait for it, then diagnose the server.
if [ -n "$domain" ]; then
	say "waiting for the certificate of $domain"
	tries=0
	# An IP address needs no resolving; a domain may not point here yet.
	resolve="--resolve=$domain:443:127.0.0.1"
	[ -z "$ip" ] || resolve="--noproxy=*"
	until curl -fsk --max-time 10 "$resolve" "https://$domain/-/healthz" >/dev/null 2>&1; do
		tries=$((tries + 1))
		[ "$tries" -lt 6 ] || break
		sleep 5
	done
fi
echo
"$bin_path" doctor --data "$data_dir" ||
	say "fix what fails above, then run sudo $bin_path doctor again"
echo
as_cahut domain list | awk 'NR > 1 { print "  https://" $2 "/" }'
[ -n "$domain" ] || domain=$(as_cahut domain list | awk 'NR == 2 { print $2 }')
[ -n "$domain" ] || die "this server holds no Cahut yet: run the script again with --domain"
# A Cahut that has nobody yet starts in the browser, from a single-use
# link. See docs/adr/0025-first-start-in-the-browser.md. The commands name
# the binary by its full path: sudo searches only its secure_path, which
# leaves out /usr/local/bin on AlmaLinux.
echo
people=$(as_cahut user list --domain "$domain" 2>/dev/null | awk 'NR > 1' | wc -l)
if [ "$people" -eq 0 ] && link=$(as_cahut setup --domain "$domain" 2>/dev/null); then
	printf '%s\n\n' "$link"
	cat <<EOF
It creates your account, sets the house style of your organization, and
connects your first AI. For a new link: sudo -u cahut $bin_path setup
EOF
else
	echo "The Cahut of $domain has people already: sign in at https://$domain/console/"
fi
cat <<EOF

Cahut updates itself every day ($updates: see CAHUT_UPDATES in
$env_file). To update now, or to see what is new:
       sudo $bin_path update --check

The service logs go to the journal: journalctl -u cahut
EOF
