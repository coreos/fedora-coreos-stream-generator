all:
	go build .
.PHONY: all

test:
	./test.sh
.PHONY: test
