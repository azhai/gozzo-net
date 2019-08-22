#!/bin/bash

GOARCH=amd64
GOOS=$(uname -s | tr [A-Z] [a-z])
APPNAME="proxy"
RELEASE="-s -w"
if [ "$GOOS" == "darwin" ]; then
    GOBIN="/usr/local/bin/go"
    UPX=""
else
    GOBIN="/usr/local/go/bin/go"
    UPX="/usr/bin/upx"
fi

rm -f "$APPNAME"
$GOBIN build -mod=vendor -ldflags="$RELEASE" -o "$APPNAME" .
chmod +x "$APPNAME"

if [ -e "$UPX" ]; then
    $UPX "$APPNAME"
fi
