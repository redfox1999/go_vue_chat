.PHONY: all build run test clean fmt vet build-linux build-windows build-macos

APP_NAME := api-server
CMD_DIR := cmd

all: build

build:
	mkdir -p bin/
	go build -o bin/$(APP_NAME) ./$(CMD_DIR)

run:
	go run ./$(CMD_DIR)

test:
	go test -race -v ./...

clean:
	rm -rf bin/


fmt:
	go fmt ./...

vet:
	go vet ./...

build-linux:
	GOOS=linux GOARCH=amd64 go build -o bin/$(APP_NAME)-linux ./$(CMD_DIR)

build-windows:
	GOOS=windows GOARCH=amd64 go build -o bin/$(APP_NAME)-windows.exe ./$(CMD_DIR)

build-macos:
	GOOS=darwin GOARCH=amd64 go build -o bin/$(APP_NAME)-macos ./$(CMD_DIR)

build-all: build-linux build-windows build-macos

run-dev:
	go run ./$(CMD_DIR)
