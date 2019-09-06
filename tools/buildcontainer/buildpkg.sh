#!/bin/bash

POSITIONAL=()
PACKAGES=()
while [[ $# -gt 0 ]]; do
	case "$1" in
		"-i")
			if [[ $# -lt 2 ]]; then
				echo "Trailing install option"
				exit 1
			fi
			PACKAGES+=("$2")
			shift
			shift
			;;
		*)
			POSITIONAL+=("$1") # save it in an array for later
			shift
			;;
	esac
done
if [[ ${#PACKAGES[@]} != 0 ]]; then
	echo "Installing prerequisites"
	pacman --noconfirm -S "${PACKAGES[@]}" || exit 1
fi

echo "Building package"
pushd /pkgroot/
su build -c makepkg || exit 1

echo "Collecting resulting packages"
mv *.pkg.tar.* /out/ || exit 1
popd
