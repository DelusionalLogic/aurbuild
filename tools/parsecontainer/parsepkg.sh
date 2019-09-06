#!/bin/bash

# source a file and fail if it does not succeed
source_safe() {
	local shellopts=$(shopt -p extglob)
	shopt -u extglob

	if ! source "$@"; then
		error "$(gettext "Failed to source %s")" "$1"
		exit 1
	fi

	eval "$shellopts"
}

source_safe "/pkgroot/PKGBUILD"

deparr=$(for f in "${depends[@]}"; do printf '%s' "$f" | jq -R -s .; done | jq -s .)
mkdeparr=$(for f in "${makedepends[@]}"; do printf '%s' "$f" | jq -R -s .; done | jq -s .)
ckdeparr=$(for f in "${checkdepends[@]}"; do printf '%s' "$f" | jq -R -s .; done | jq -s .)

cat <<EOF >/out/info.json
{
	"name": "$pkgname",
	"version": "$pkgver",
	"depends": {
		"install": $deparr,
		"build": $mkdeparr,
		"check": $ckdeparr
	}
}
EOF

if [[ ! -z "$FUID" ]] && [[ ! -z "$FGID" ]]; then
	chown $FUID:$FGID /out/info.json
fi
