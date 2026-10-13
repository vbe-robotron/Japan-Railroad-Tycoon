.PHONY: build test run-server run-cli lint clean

build:
	go build ./...

test:
	go test ./...

run-server:
	go run ./cmd/server

run-cli:
	go run ./cmd/cli

lint:
	go vet ./...

clean:
	rm -rf bin/
