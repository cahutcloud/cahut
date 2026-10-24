#!/bin/sh
# Installs Cahut on this blank machine with install.sh, then checks what an
# administrator relies on: the service and the timer of updates, the health
# check, cahut doctor, the link to the first start, an update, and the
# rollback of a binary that does not start. Run it as root, with install.sh
# and the binaries of two versions of Cahut for this machine:
#
#   sudo sh test/install/run.sh install.sh cahut-a cahut-b
#
# CI runs it on the Ubuntu machine of GitHub; test/install/lima.sh runs it
# in blank virtual machines. See docs/plan-lot-7.md.

set -eu

script=${1:?give install.sh}
a=${2:?give the binary to install}
b=${3:?give the binary to update to}
bin=/usr/local/bin/cahut

step() { printf '\n== %s\n' "$*"; }
fail() { printf 'install test: %s\n' "$*" >&2; exit 1; }
version() { "$1" version | awk '{ print $2 }'; }
# answers VERSION succeeds when the service answers on loopback with
# VERSION, every Cahut open.
answers() {
	h=$(curl -fsS --max-time 5 http://127.0.0.1/-/healthz) || return 1
	echo "$h" | grep -q "\"version\":\"$1\"" && echo "$h" | grep -q '"ok":true'
}

step "install.sh refuses an account named cahut that a person signs in with"
useradd --create-home --shell /bin/sh cahut
out=$(sh "$script" --binary "$a" --domain cahut.test --staging 2>&1) && fail "install.sh took the account cahut for the service"
echo "$out"
echo "$out" | grep -q "is not a system account" || fail "install.sh did not say why it refused the account cahut"
userdel --remove cahut 2>/dev/null || true
! id cahut >/dev/null 2>&1 || fail "cannot delete the account cahut"

started=$(date +%s)

step "install.sh --binary $(version "$a")"
out=$(sh "$script" --binary "$a" --domain cahut.test --staging 2>&1) || {
	echo "$out"
	fail "install.sh failed"
}
echo "$out" | sed -E 's/(setup#)[a-z0-9]+/\1<code>/'
echo "$out" | grep -q '/console/setup#' || fail "install.sh printed no link to the first start"
systemctl is-active --quiet cahut || fail "the cahut service does not run"
systemctl is-enabled --quiet cahut || fail "the cahut service does not start at boot"
systemctl is-active --quiet cahut-update.timer || fail "the timer of updates does not run"
systemctl is-active --quiet cahut-update.path || fail "the path unit of updates does not run"
answers "$(version "$a")" || fail "the service does not answer with $(version "$a")"
installed=$(date +%s)

step "cahut doctor"
"$bin" doctor >"/tmp/doctor.txt" 2>&1 || true
cat /tmp/doctor.txt
for line in "ok    Service" "ok    Data" "ok    Ports" "ok    Updates"; do
	grep -q "^$line" /tmp/doctor.txt || fail "cahut doctor does not say: $line"
done

step "cahut update --binary $(version "$b")"
"$bin" update --binary "$b" || fail "the update failed"
answers "$(version "$b")" || fail "the service does not answer with $(version "$b") after the update"

step "cahut update --binary, with a binary that does not start"
broken=$(mktemp)
printf '#!/bin/sh\n[ "$1" = version ] && { echo "cahut 9.9.9-broken (test)"; exit 0; }\necho broken on purpose >&2\nexit 1\n' >"$broken"
chmod 0755 "$broken"
"$bin" update --binary "$broken" || fail "the update with a broken binary did not roll back"
grep -q '"result": "rolled_back"' /var/lib/cahut-update/status.json || fail "the update was not rolled back"
answers "$(version "$b")" || fail "the service does not answer with $(version "$b") after the rollback"
rm -f "$broken"

step "done"
printf 'install.sh took %ss; the whole test %ss.\n' "$((installed - started))" "$(($(date +%s) - started))"
