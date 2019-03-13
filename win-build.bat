@ECHO OFF

del proxy.exe
go build -mod=vendor -ldflags="-s -w" -o proxy.exe main.go

PAUSE
