#!/bin/bash

GOARCH=amd64
GOOS=$(uname -s | tr [A-Z] [a-z])
RELEASE="-s -w"
if [ "$GOOS" == "darwin" ]; then
    GOBIN="/usr/local/bin/go"
    UPX=""
else
    GOBIN="/usr/local/go/bin/go"
    UPX="/usr/bin/upx"
fi

rm -f proxy server
$GOBIN build -mod=vendor -ldflags="$RELEASE" -o proxy ./examples/proxy/
$GOBIN build -mod=vendor -ldflags="$RELEASE" -o server ./examples/server/
chmod +x proxy server

if [ -e "$UPX" ]; then
    $UPX proxy server
fi
