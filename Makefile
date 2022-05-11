VERSION := $(shell ./version.sh)

.PHONY: all build gen deps

all: build

build: deps
	go build -o build/elc -ldflags="-X 'github.com/madridianfox/elc/src.Version=${VERSION}'" main.go

deps:
	go get

install:
	mkdir -p /opt/elc
	sudo cp ./build/elc /opt/elc/elc-v${VERSION}
	sudo ln -sf /opt/elc/elc-v${VERSION} /usr/local/bin/elc