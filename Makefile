.PHONY: build run test lint clean

build:
	go build -o bin/global-ranks ./cmd/global-ranks

run: build
	./bin/global-ranks

test:
	go test -v -race -count=1 ./...

lint:
	go vet ./...
	staticcheck ./...

clean:
	rm -rf bin/