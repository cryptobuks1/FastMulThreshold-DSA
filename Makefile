# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

.PHONY: all gsmpc bootnode clean fmt gsmpc-client

all:
	./build.sh gsmpc bootnode gsmpc-client
	cp cmd/conf.toml bin/cmd
	@echo "Done building."

gsmpc:
	./build.sh gsmpc
	@echo "Done building."

bootnode:
	./build.sh bootnode
	@echo "Done building."

gsmpc-client:
	./build.sh gsmpc-client
	@echo "Done building."

clean:
	rm -fr bin/cmd/*
	./build.sh gsmpc-test-clean

fmt:
	./gofmt.sh

gsmpc-test:
	./build.sh gsmpc-test

gsmpc-test-short:
	./build.sh gsmpc-test-short
