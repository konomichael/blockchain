VERSION = $(shell git describe --tags --always --dirty)
COMMIT = $(shell git rev-parse HEAD)
DATE = $(shell date -u '+%Y-%m-%d_%I:%M:%S%p')

build-chain:
	go build -o bin/chain -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)" ./cmd/blockchain

build-wallet:
	go build -o bin/wallet -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)" ./cmd/wallet

code-check:
	golangci-lint run