#!/bin/sh
# The forced command of the SSH key that GitHub Actions uses to upload the
# sites (see README.md, Deployment). The key can only run rsync in server
# mode, receiving into the document root of one of the sites: no shell, no
# download, no other folder.
#
# On the server, it sits next to the key in ~/.ssh/authorized_keys:
#   command="$HOME/.ssh/deploy-command.sh",restrict ssh-ed25519 AAAA... sites-deploy

set -f
# shellcheck disable=SC2086 # split the command into words, without globbing
set -- $SSH_ORIGINAL_COMMAND

refuse() {
	echo "This key only uploads the sites with rsync: $1." >&2
	exit 1
}

[ "$1" = rsync ] && [ "$2" = --server ] || refuse "not an rsync upload"
for arg; do
	case "$arg" in
	--sender) refuse "no download" ;;
	--log-file* | --temp-dir* | -T | --backup-dir* | --partial-dir* | --compare-dest* | --copy-dest* | --link-dest*)
		refuse "no option that writes elsewhere" ;;
	esac
	last="$arg"
done
case "$last" in
httpdocs/ | docs.cahut.cloud/) ;;
*) refuse "no folder other than httpdocs/ and docs.cahut.cloud/" ;;
esac

exec "$@"
