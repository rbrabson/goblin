include Configfile

.PHONY: build
build: build-linux build-mac-amd build-mac-arm build-windows

build-windows: export GOOS=windows
build-windows: export GOARCH=amd64
build-windows: export GO111MODULE=on
build-windows: export GOPROXY=$(MOD_PROXY_URL)
build-windows:
	$(GO) build -v --ldflags="-w -X main.Version=$(VERSION) -X main.Revision=$(REVISION) -X main.BotName=$(BOT_NAME)" \
		-o bin/windows/amd64/goblin cmd/goblin/main.go  # windows

build-linux: export GOOS=linux
build-linux: export GOARCH=amd64
build-linux: export CGO_ENABLED=0
build-linux: export GO111MODULE=on
build-linux: export GOPROXY=$(MOD_PROXY_URL)
build-linux:
	$(GO) build -v --ldflags="-w -X main.Version=$(VERSION) -X main.Revision=$(REVISION) -X main.BotName=$(BOT_NAME)" \
		-o bin/linux/amd64/goblin cmd/goblin/main.go  # linux

build-mac-amd: export GOOS=darwin
build-mac-amd: export GOARCH=amd64
build-mac-amd: export CGO_ENABLED=0
build-mac-amd: export GO111MODULE=on
build-mac-amd: export GOPROXY=$(MOD_PROXY_URL)
build-mac-amd:
	$(GO) build -v --ldflags="-w -X main.Version=$(VERSION) -X main.Revision=$(REVISION) -X main.BotName=$(BOT_NAME)" \
		-o bin/macos/amd64/goblin cmd/goblin/main.go  # mac osx intel chip

build-mac-arm: export GOOS=darwin
build-mac-arm: export GOARCH=arm64
build-mac-arm: export CGO_ENABLED=0
build-mac-arm: export GO111MODULE=on
build-mac-arm: export GOPROXY=$(MOD_PROXY_URL)
build-mac-arm:
	$(GO) build -v --ldflags="-w -X main.Version=$(VERSION) -X main.Revision=$(REVISION) -X main.BotName=$(BOT_NAME)" \
		-o bin/macos/arm64/goblin cmd/goblin/main.go  # mac osx arm chip

.PHONY: clean
clean::
	echo "--> cleaning..."
	rm -rf vendor
	go clean ./...