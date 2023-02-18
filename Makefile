VERSION = $(shell git describe --tags --always --dirty)
COMMIT = $(shell git rev-parse HEAD)
DATE = $(shell date -u '+%Y-%m-%d_%I:%M:%S%p')

build:
	go build -o bin/blockchain -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

code-check:
	golangci-lint run