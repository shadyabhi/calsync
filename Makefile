# Default target
all: build

# Build target
build:
	@echo "Building project..."
	go build

run: build
	@echo "Running calsync..."
	./calsync

test:
	@echo "Running tests..."
	go test -race -count=1 -v ./...

# Clean target (optional)
clean:
	@echo "Cleaning up..."
	rm -r calcync

.PHONY: all build clean
