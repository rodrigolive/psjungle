# Default target
.DEFAULT_GOAL := build

# Build psjungle binary
build:
	go build -o psjungle ./cmd/psjungle

# Clean built binary
clean:
	rm -f psjungle

# Install dependencies
deps:
	go mod tidy

# Run tests
test:
	go test -v ./...

# Build and install binary to GOPATH
install: build
	go install ./cmd/psjungle

# Help
help:
	@echo "Available targets:"
	@echo "  build   - Build psjungle binary (default)"
	@echo "  clean   - Remove built binary"
	@echo "  deps    - Install/update dependencies"
	@echo "  test    - Run tests"
	@echo "  install - Build and install binary to GOPATH"
	@echo "  help    - Show this help"

.PHONY: build clean deps test install help