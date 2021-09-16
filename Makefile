.PHONY: all
all: fedora-coreos-stream-generator

fedora-coreos-stream-generator: main.go go.mod Makefile
	go build .

.PHONY: test
test:
	./test.sh
