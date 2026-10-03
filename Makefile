VERSION := $(shell ./version.sh)

.PHONY: all build gen deps test coverage e2e

all: build

build: deps
	go build -o build/elc -ldflags="-X 'github.com/ensi-platform/elc/core.Version=${VERSION}'" main.go

deps:
	go get

install:
	sudo mkdir -p /opt/elc
	sudo cp ./build/elc /opt/elc/elc-v${VERSION}
	sudo ln -sf /opt/elc/elc-v${VERSION} /usr/local/bin/elc

test:
	go test -v ./...

coverage:
	go test -coverprofile=coverage.out -v ./...
	go tool cover -html=coverage.out

e2e:
	./e2e/run.sh

