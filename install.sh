#!/bin/sh
# Install or update Cahut on a Linux server with systemd.
#
#   sudo sh install.sh --domain cahut.example.com [--email you@example.com]
#
# The script downloads a release from GitHub and checks its checksum, or
# takes a binary you give it with --binary. It then creates the cahut system
# user, the data directory /var/lib/cahut, the settings file
# /etc/cahut/cahut.env and the cahut service, declares the Cahut of
# --domain, opens ports 80 and 443 in firewalld or ufw when one is active,
# and starts the service.
#
# Options:
#   --domain DOMAIN   domain of a Cahut to declare on this server, such as
#                     cahut.example.com; it must point to this server. Run
#                     the script again with another domain to add a Cahut.
#   --email EMAIL     contact of the Let's Encrypt account (optional)
#   --staging         use Let's Encrypt's staging environment, for tests
#   --version TAG     release to install, such as v0.1.0 (default: latest)
#   --binary PATH     install this cahut binary instead of downloading one
#
# Environment:
#   CAHUT_REPO        GitHub repository of the releases (default: getcahut/cahut)
#   GITHUB_TOKEN      token to read the releases of a private repository
#
# Run it again to update Cahut or change its settings.

set -eu

repo="${CAHUT_REPO:-getcahut/cahut}"
domain=""
email=""
ca=""
tag="latest"
binary=""

bin_path=/usr/local/bin/cahut
data_dir=/var/lib/cahut
conf_dir=/etc/cahut
env_file=$conf_dir/cahut.env
unit_file=/etc/systemd/system/cahut.service

say() { printf 'cahut: %s\n' "$*"; }
die() { printf 'cahut: %s\n' "$*" >&2; exit 1; }

while [ $# -gt 0 ]; do
	case "$1" in
	--domain) domain="${2:?--domain needs a value}"; shift 2 ;;
	--email) email="${2:?--email needs a value}"; shift 2 ;;
	--staging) ca=staging; shift ;;
	--version) tag="${2:?--version needs a value}"; shift 2 ;;
	--binary) binary="${2:?--binary needs a value}"; shift 2 ;;
	-h | --help) sed -n '2,/^$/s/^# \{0,1\}//p' "$0"; exit 0 ;;
	*) die "unknown option $1 (see --help)" ;;
	esac
done

[ "$(id -u)" -eq 0 ] || die "run this script as root, for instance with sudo"
command -v systemctl >/dev/null || die "systemd is required"
if [ -z "$domain" ] && [ ! -f "$env_file" ]; then
	die "give the domain of your first Cahut with --domain"
fi
if [ -n "$domain" ]; then
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
	for file in "$archive" checksums.txt; do
		url=$(asset_url "$file")
		[ -n "$url" ] || die "the release has no $file"
		fetch "$url" "$tmp/$file" application/octet-stream || die "cannot download $file"
	done
	(cd "$tmp" && grep " $archive\$" checksums.txt | sha256sum -c --quiet -) || die "checksum mismatch for $archive"
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

if [ ! -f "$env_file" ] || [ -n "$email$ca" ]; then
	say "writing $env_file"
	umask 027
	cat >"$env_file" <<EOF
# Settings of the cahut service. Run install.sh again to change them.
CAHUT_DATA=$data_dir
CAHUT_ACME_EMAIL=$email
CAHUT_ACME_CA=$ca
EOF
	chgrp cahut "$env_file"
	umask 022
fi

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

[Install]
WantedBy=multi-user.target
EOF

install -m 0755 "$tmp/cahut" "$bin_path"
if command -v restorecon >/dev/null; then
	restorecon "$bin_path" "$unit_file" || true # SELinux labels of the new files
fi

# as_cahut runs a cahut command as the service user, on its data.
as_cahut() { runuser -u cahut -- "$bin_path" "$@" --data "$data_dir"; }

if [ -n "$domain" ] && ! as_cahut domain list | awk 'NR > 1 { print $2 }' | grep -qx "$domain"; then
	say "declaring the Cahut of $domain"
	as_cahut domain add "$domain" >/dev/null
fi

if systemctl is-active --quiet firewalld 2>/dev/null; then
	say "opening ports 80 and 443 in firewalld"
	firewall-cmd --quiet --permanent --add-service=http --add-service=https
	firewall-cmd --quiet --reload
elif command -v ufw >/dev/null && ufw status 2>/dev/null | grep -q '^Status: active'; then
	say "opening ports 80 and 443 in ufw"
	ufw allow 80/tcp >/dev/null
	ufw allow 443/tcp >/dev/null
fi

systemctl daemon-reload
systemctl enable --quiet cahut
systemctl restart cahut
sleep 2
systemctl is-active --quiet cahut || die "the service did not start: see journalctl -u cahut"

say "$("$bin_path" version) is running"
as_cahut domain list | awk 'NR > 1 { print "  https://" $2 "/" }'
[ -n "$domain" ] || domain=$(as_cahut domain list | awk 'NR == 2 { print $2 }')
[ -n "$domain" ] || die "this server holds no Cahut yet: run the script again with --domain"
# The commands name the binary by its full path: sudo searches only its
# secure_path, which leaves out /usr/local/bin on AlmaLinux.
cat <<EOF

Next steps for the Cahut of $domain:

  1. Create your account, as an administrator:
       sudo -u cahut $bin_path user create --domain $domain --handle you --name "Your Name" --admin
  2. Create a token for your AI:
       sudo -u cahut $bin_path token create --domain $domain --user you --name "Claude Code"
  3. Add the Cahut to Claude Code:
       claude mcp add --transport http cahut https://$domain/mcp --header "Authorization: Bearer <token>"

The service logs go to the journal: journalctl -u cahut
EOF
