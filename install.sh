#!/bin/sh
# Copyright 2019 the Deno authors. All rights reserved. MIT license.
# Copyright 2022 the Backpack authors. All rights reserved. MIT license.
# TODO(everyone): Keep this script simple and easily auditable.

set -e

# assumes bin name same as project name
project="teller"
bin_name="teller"
repo="tellerops/teller"

if [ "$OS" = "Windows_NT" ]; then
	target="x86_64-windows"
else
	case $(uname -sm) in
	"Darwin x86_64") target="x86_64-macos" ;;
	"Darwin arm64") target="aarch64-macos" ;;
	*) target="x86_64-linux" ;;
	esac
fi

if [ $# -eq 0 ]; then
	uri="https://github.com/${repo}/releases/latest/download/${project}-${target}.tar.xz"
else
	uri="https://github.com/${repo}/releases/download/${1}/${project}-${target}.tar.xz"
fi

install="${PROJ_INSTALL:-$HOME/.$project-bin}"
bin_dir="$install"
exe="$bin_dir/$bin_name"

if [ ! -d "$bin_dir" ]; then
	mkdir -p "$bin_dir"
fi

curl --fail --location --progress-bar --output "$exe.tar.xz" "$uri"
tar zxf "$exe.tar.xz" -C "$bin_dir" --strip-components 1 
chmod +x "$exe"
rm "$exe.tar.xz"

echo "$project was installed successfully to $exe"
if command -v $exe >/dev/null; then
	echo "Run '$exe --help' to get started"
else
	case $SHELL in
	/bin/zsh) shell_profile=".zshrc" ;;
	*) shell_profile=".bashrc" ;;
	esac
	echo "Manually add the directory to your \$HOME/$shell_profile (or similar)"
	echo "  export PROJ_INSTALL=\"$install\""
	echo "  export PATH=\"\$PROJ_INSTALL:\$PATH\""
	echo "Run '$exe --help' to get started"
fi
