# Default target
all: build test

# Build target
build:
	@echo "Building project..."
	go build

run: build
	@echo "Running calsync..."
	./calsync

test:
	@echo "Running tests..."
	go test -race -count=1 -v ./... -coverprofile=./cover.out

goreleaser:
        goreleaser release --snapshot --clean

clean:
	@echo "Cleaning up..."
	rm -r calcync

.PHONY: all build clean
