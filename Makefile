.PHONY: build test test-integration

build:
	go build -o trivia .

test:
	go fmt ./... && go vet ./... && go test -v ./...

test-integration:
	go test -v -tags=integration ./...