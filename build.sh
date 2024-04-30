#!/bin/sh
OSes="linux freebsd"
ARCHes="arm64 amd64"

for OS in $OSes; do
    for ARCH in $ARCHes; do
        GOOS="$OS" GOARCH="$ARCH" go build -o scpez-"$OS-$ARCH" main.go
    done
done

GOOS=darwin GOARCH=amd64 go build -o scpez-mac main.go