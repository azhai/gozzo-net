@ECHO OFF

del proxy.exe server.exe
go.exe build -mod=vendor -ldflags="-s -w" -o proxy.exe ./examples/proxy/
go.exe build -mod=vendor -ldflags="-s -w" -o server.exe ./examples/server/

PAUSE
