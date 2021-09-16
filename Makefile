VERSION=$(shell git describe --dirty --always)

.PHONY: all
all: fedora-coreos-stream-generator

fedora-coreos-stream-generator: main.go go.mod Makefile
	go build -ldflags "$(GLDFLAGS) -X main.Version=$(VERSION)" .

.PHONY: test
test:
	./test.sh
