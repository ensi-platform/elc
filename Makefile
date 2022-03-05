.PHONY: all build gen deps

all: build

build: deps
	go build -o build/elc main.go

deps:
	go get

install:
	mkdir -p /opt/elc
	cp ./build/elc /opt/elc/elc-dev
	ln -s /opt/elc/elc-dev /usr/local/bin/elc