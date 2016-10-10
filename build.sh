#!/bin/bash

function build()
{
	TARGET="$1"

	echo "Building icepacker for $TARGET..."
	GOOS="${TARGET%-*}" GOARCH="${TARGET##*-}" go build \
		-o release/ipack-$TARGET \
		cmd/ipack/main.go
}

for target in \
	darwin-386 darwin-amd64 \
	freebsd-386 freebsd-amd64 \
	linux-386 linux-amd64  \
	netbsd-386 netbsd-amd64 \
	openbsd-386 openbsd-amd64 \
	windows-386 windows-amd64; do
	build $target || exit 1
done