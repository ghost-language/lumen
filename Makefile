GOPATH:=$(shell go env GOPATH)

.PHONY: build compile clean

build: clean
	GOOS=darwin go build -o dist/lumen cmd/*.go

compile: clean
	GOOS=darwin go build -tags static -ldflags "-s -w" -o dist/lumen cmd/*.go

clean:
	@rm -rf dist/lumen