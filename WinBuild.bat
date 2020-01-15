@ECHO OFF

git.exe checkout master
git.exe pull --all

del proxy.exe relay.exe server.exe
go.exe build -ldflags="-s -w" -o proxy.exe ./cmd/proxy
go.exe build -ldflags="-s -w" -o relay.exe ./cmd/relay
go.exe build -ldflags="-s -w" -o server.exe ./cmd/server
