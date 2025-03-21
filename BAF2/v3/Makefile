.PHONY: build clean test setup wallet run-client run-device help

# Project variables
PROJECT_NAME := auth-framework
BIN_DIR := bin
CMD_DIR := cmd
CONFIG_DIR := config

# Build targets
build:
	@echo "Building authentication framework..."
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/authcli $(CMD_DIR)/authcli/main.go

clean:
	@echo "Cleaning up..."
	@rm -rf $(BIN_DIR)
	@rm -f *-tgt.json *-serviceticket-*.json client1-session-*.json
	@echo "Cleaned build artifacts"

test:
	@echo "Running tests..."
	@go test -v ./internal/... ./pkg/...

setup:
	@echo "Setting up environment..."
	@mkdir -p $(CONFIG_DIR) keys wallet sessions
	@./scripts/setup.sh

wallet:
	@echo "Initializing wallet..."
	@./scripts/init-wallet.sh

# Register client and device
register-client:
	@$(BIN_DIR)/authcli register-client --client-id client1

register-device:
	@$(BIN_DIR)/authcli register-device --device-id device1 --capabilities temperature,humidity,pressure

# Run authentication flow
authenticate:
	@$(BIN_DIR)/authcli authenticate --client-id client1 --device-id device1

access-device:
	@$(BIN_DIR)/authcli access-device --client-id client1 --device-id device1

get-device-data:
	@$(BIN_DIR)/authcli get-device-data --device-id device1

close-session:
	@$(BIN_DIR)/authcli close-session --client-id client1 --device-id device1

list-sessions:
	@$(BIN_DIR)/authcli list-sessions

# Complete authentication flow
auth-flow: register-client register-device authenticate access-device get-device-data

# Help information
help:
	@echo "Authentication Framework Makefile"
	@echo "--------------------------------"
	@echo "Available targets:"
	@echo "  build            - Build the authentication CLI binary"
	@echo "  clean            - Remove build artifacts"
	@echo "  test             - Run tests"
	@echo "  setup            - Set up environment (create directories, config)"
	@echo "  wallet           - Initialize wallet with identity"
	@echo "  register-client  - Register a client with AS"
	@echo "  register-device  - Register a device with ISV"
	@echo "  authenticate     - Authenticate client with AS and get TGT"
	@echo "  access-device    - Request access to device"
	@echo "  get-device-data  - Get device data"
	@echo "  close-session    - Close an active session"
	@echo "  list-sessions    - List active sessions"
	@echo "  auth-flow        - Run complete authentication flow"
	@echo "  help             - Display this help information"

.DEFAULT_GOAL := help
