BIN = snake
SRC = $(shell find . -name '*.go')

all: test build

build: bin/$(BIN)

test:
	@go test -v ./...

clean:
	@rm -r ./bin

bin/$(BIN): $(SRC)
	@mkdir -p ./bin
	@go mod tidy
	@go build -ldflags="-s -w" -o ./bin/$(BIN) .

.PHONY: all build test clean
