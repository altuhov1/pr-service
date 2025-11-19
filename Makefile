
BINARY_NAME=httpBack
CMD_PATH=./cmd/httpBack
BUILD_DIR=./bin
GO=go
run:
	$(GO) run $(CMD_PATH)

build:
	$(GO) build -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_PATH)
clean:
	rm -rf $(BUILD_DIR)

test:
	$(GO) test ./...

deps:
	$(GO) mod tidy && $(GO) mod download

.PHONY: help run build clean test deps