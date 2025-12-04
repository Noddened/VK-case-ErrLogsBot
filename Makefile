.PHONY: clean build run

build:
	go build -o ./bin/ErrLogsBot ./cmd/ErrLogsBot/main.go 

clean:
	go clean
	rm -rf ./bin && mkdir ./bin
	rm -f payload/logs.log

run: build
	./bin/ErrLogsBot