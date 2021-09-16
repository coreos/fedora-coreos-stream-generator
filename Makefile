all: fedora-coreos-stream-generator
.PHONY: all

fedora-coreos-stream-generator: main.go go.mod Makefile
	go build .

test:
	./test.sh
.PHONY: test
