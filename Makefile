BINARY=mentor
SOURCEDIR=.
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')
GOOS := $(shell uname | tr '[:upper:]' '[:lower:]')
GOARCH := amd64

.DEFAULT_GOAL := build

all: test build install clean

test:
	GOOS=${GOOS} GOARCH=${GOARCH} go test main.go

build:
	GOOS=${GOOS} GOARCH=${GOARCH} go build -o bin/${GOOS}/${BINARY} main.go

clean:
	rm -rf bin/*
