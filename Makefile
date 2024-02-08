BINARY_NAME=gosymex
BINARY_PATH=./bin/$(BINARY_NAME)

build:
	mkdir -p ./bin
	go build -o $(BINARY_PATH) main.go

run: build
	$(BINARY_PATH)

clean:
	go clean
	rm -rf ./bin

test:
	go test -v ./...

deps:
	go get -u
