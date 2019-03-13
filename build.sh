#!/bin/bash

GOARCH=amd64
GOOS=$(uname -s | tr [A-Z] [a-z])
APPNAME="proxy"
RELEASE="-s -w"
if [ "$GOOS" == "darwin" ]; then
    GOBUILD="/usr/local/bin/go build -mod=vendor"
    UPX=""
else
    GOBUILD="/usr/bin/go build -mod=vendor"
    UPX="/usr/bin/upx"
fi

rm -f "$APPNAME"
$GOBUILD -ldflags="$RELEASE" -o "$APPNAME" *.go
chmod +x "$APPNAME"

if [ -e "$UPX" ]; then
    $UPX "$APPNAME"
fi
