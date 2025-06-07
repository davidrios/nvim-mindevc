#!/bin/bash
set -e

if test $(uname -s) != "Linux"; then
	echo only linux supported
	exit 1
fi

ARCH=$(uname -m)
ZIG_VER="0.14.1"
ZIG_NAME="zig-$ARCH-linux-$ZIG_VER.tar.xz"

mkdir -p tmp/zig

test -f "tmp/$ZIG_NAME" || (cd tmp && curl -LO "https://ziglang.org/download/$ZIG_VER/$ZIG_NAME")

MYPATH="$PWD/tmp/zig:$PATH"

test "$(PATH="$MYPATH" zig version)" == "$ZIG_VER" || tar xJf "tmp/$ZIG_NAME" -C tmp/zig --strip-components=1

(cd setup/testdata/tools && PATH="$MYPATH" make)
