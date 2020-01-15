RELEASE=-s -w
UPXBIN=/usr/local/bin/upx
GOBIN=/usr/local/bin/go
GOOS=$(shell uname -s | tr [A-Z] [a-z])
GOARGS=GOARCH=amd64 CGO_ENABLED=0
GOBUILD=$(GOARGS) $(GOBIN) build -ldflags="$(RELEASE)"

.PHONY: all
all: clean build
build:
	@echo Compile proxy ...
	GOOS=$(GOOS) $(GOBUILD) -o proxy ./cmd/proxy
	@echo Build proxy success.
	@echo Compile relay ...
	GOOS=$(GOOS) $(GOBUILD) -o relay ./cmd/relay
	@echo Build relay success.
	@echo Compile server ...
	GOOS=$(GOOS) $(GOBUILD) -o server ./cmd/server
	@echo Build server success.
clean:
	rm -f proxy relay server
	@echo Clean all.
upx: build command
	$(UPXBIN) proxy relay server
upxx: build command
	$(UPXBIN) --ultra-brute proxy relay server
vend:
	GOOS=$(GOOS) $(GOBUILD) -mod=vendor -o proxy ./cmd/proxy
	GOOS=$(GOOS) $(GOBUILD) -mod=vendor -o relay ./cmd/relay
	GOOS=$(GOOS) $(GOBUILD) -mod=vendor -o server ./cmd/server
