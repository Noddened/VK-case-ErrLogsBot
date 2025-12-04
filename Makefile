.PHONY: clean

all := run


build:
	go build -o ./bin/bot_app.bin ./cmd/ErrLogsBot/main.go 

clean:
	go clean
	rm -rf ./bin && mkdir ./bin
	rm payload/logs.log

run: build
	./bin/bot_app.bin