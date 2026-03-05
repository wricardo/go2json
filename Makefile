# Makefile for go2json

.DEFAULT_GOAL := install

.PHONY: build install test

install:
	go install ./cmd/go2json

test:
	go build -o /dev/null ./cmd/go2json/
	go test -v ./...

build-delve:
	go build -gcflags "all=-N -l" -o ./bin/go2json ./cmd/go2json
