@ECHO OFF

del proxy.exe
go.exe build -mod=vendor -ldflags="-s -w" -o proxy.exe .

PAUSE
