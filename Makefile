APP_NAME := dialectApi
SRC := ./cmd/server

.PHONY: build run clean fmt lint

build:
	go build -o $(APP_NAME) $(SRC)

run:
	go run cmd/server/main.go

