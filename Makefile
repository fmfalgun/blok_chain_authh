.PHONY: help install-deps network-up network-down channel-create deploy-cc test clean monitoring-up monitoring-down

help: ## Display this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

install-deps: ## Install project dependencies
	@echo "Installing Go dependencies..."
	cd chaincodes/as-chaincode && go mod download
	cd chaincodes/tgs-chaincode && go mod download
	cd chaincodes/isv-chaincode && go mod download
	@echo "Installing Node.js dependencies..."
	cd blockchain-auth-framework && npm install

network-up: ## Start the Hyperledger Fabric network
	@echo "Starting Hyperledger Fabric network..."
	cd network && ./scripts/network.sh up

network-down: ## Stop the Hyperledger Fabric network
	@echo "Stopping Hyperledger Fabric network..."
	cd network && ./scripts/network.sh down

channel-create: ## Create the authentication channel
	@echo "Creating authentication channel..."
	cd network && ./scripts/network.sh createChannel

deploy-cc: ## Deploy all chaincodes
	@echo "Deploying chaincodes..."
	cd network && ./scripts/deploy-chaincode.sh as
	cd network && ./scripts/deploy-chaincode.sh tgs
	cd network && ./scripts/deploy-chaincode.sh isv

test: ## Run all tests
	@echo "Running unit tests..."
	cd tests && ./run-tests.sh unit
	@echo "Running integration tests..."
	cd tests && ./run-tests.sh integration

test-unit: ## Run unit tests only
	cd tests && ./run-tests.sh unit

test-integration: ## Run integration tests only
	cd tests && ./run-tests.sh integration

test-performance: ## Run performance tests
	cd tests && ./run-tests.sh performance

monitoring-up: ## Start monitoring stack (Prometheus, Grafana, Explorer)
	@echo "Starting monitoring stack..."
	cd monitoring && docker-compose -f docker-compose-monitoring.yml up -d

monitoring-down: ## Stop monitoring stack
	@echo "Stopping monitoring stack..."
	cd monitoring && docker-compose -f docker-compose-monitoring.yml down

clean: ## Clean up all generated artifacts
	@echo "Cleaning up..."
	rm -rf network/crypto-config
	rm -rf network/channel-artifacts
	rm -rf network/ledgers
	rm -rf chaincodes/*/vendor
	docker rm -f $$(docker ps -aq) 2>/dev/null || true
	docker volume prune -f
	docker network prune -f

restart: network-down network-up channel-create deploy-cc ## Restart the entire network

logs: ## Show logs for all containers
	docker-compose -f network/config/docker-compose.yaml logs -f

verify: ## Verify network and channel status
	cd network && ./scripts/verify-channel.sh

package: ## Package chaincodes
	@echo "Packaging chaincodes..."
	cd chaincodes/as-chaincode && tar czf ../../as-chaincode.tar.gz .
	cd chaincodes/tgs-chaincode && tar czf ../../tgs-chaincode.tar.gz .
	cd chaincodes/isv-chaincode && tar czf ../../isv-chaincode.tar.gz .
