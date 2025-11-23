
BINARY_NAME=httpBack
CMD_PATH=./cmd/httpBack
BUILD_DIR=./bin
E2E_SCRIPT=./scripts/e2e/e2e_test.sh
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

e2e:
	@if [ -f "$(E2E_SCRIPT)" ]; then \
		chmod +x $(E2E_SCRIPT) && \
		$(E2E_SCRIPT); \
	else \
		echo "Create scripts/e2e_test.sh with curl commands"; \
		exit 1; \
	fi
.PHONY: help run build clean test deps e2e