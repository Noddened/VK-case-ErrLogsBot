.PHONY: build run

build:
	go build -o ./bin/ErrLogsBot ./cmd/ErrLogsBot/main.go 


run: build
	./bin/ErrLogsBot