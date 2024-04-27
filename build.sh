#!/bin/sh
OSes="linux darwin freebsd"
ARCHes="arm64 amd64"

for OS in $OSes; do
    for ARCH in $ARCHes; do
        GOOS="$OS" GOARCH="$ARCH" go build -o scpez-"$OS-$ARCH" main.go
    done
done