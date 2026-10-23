#!/bin/sh
# Runs test/install/run.sh in a blank virtual machine, made by Lima from one
# of its templates, then deletes the machine:
#
#   test/install/lima.sh almalinux-10
#   test/install/lima.sh ubuntu-24.04
#
# It builds two versions of Cahut for the architecture of this computer, in
# tmp/install-test, which Lima mounts read-only in the machine. SELinux
# enforces on AlmaLinux, as on a server. make install-test runs it.

set -eu

system=${1:?give a template of Lima, such as almalinux-10 or ubuntu-24.04}
command -v limactl >/dev/null || { echo "install test: Lima is required: brew install lima" >&2; exit 1; }
root=$(cd "$(dirname "$0")/../.." && pwd)
name=cahut-test-$(printf '%s' "$system" | tr -c 'a-z0-9\n' '-')
case "$(uname -m)" in
x86_64 | amd64) arch=amd64 ;;
*) arch=arm64 ;;
esac

dir=$root/tmp/install-test
mkdir -p "$dir"
for v in 0.0.0-test.1 0.0.0-test.2; do
	(cd "$root" && CGO_ENABLED=0 GOOS=linux GOARCH=$arch go build -trimpath \
		-ldflags "-X cahut.cloud/internal/version.version=$v" -o "$dir/cahut-$v" ./cmd/cahut)
done

limactl delete -f "$name" >/dev/null 2>&1 || true
limactl start --name "$name" --tty=false "template:$system" >/dev/null
trap 'limactl delete -f "$name" >/dev/null 2>&1' EXIT
case "$system" in
almalinux*) limactl shell "$name" sudo setenforce 1 ;;
esac
limactl shell "$name" sudo sh "$root/test/install/run.sh" "$root/install.sh" \
	"$dir/cahut-0.0.0-test.1" "$dir/cahut-0.0.0-test.2"
