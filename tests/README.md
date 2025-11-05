# Testing Suite

📍 **Location**: `tests/`
🔗 **Parent**: [Main README](../README.md)

## Overview
Comprehensive testing framework covering unit tests, integration tests, and performance benchmarks.

## Test Types

### 1. Unit Tests (`unit/`)
Test individual chaincode functions in isolation.
```bash
cd tests
./run-tests.sh unit
```

### 2. Integration Tests (`integration/`)
Test complete authentication flows across all chaincodes.
```bash
./run-tests.sh integration
```

### 3. Performance Tests (`performance/`)
Benchmark throughput and latency.
```bash
./run-tests.sh performance
```

## Directory Structure
```
tests/
├── run-tests.sh              ← Test runner
├── unit/
│   └── as_chaincode_test.go
├── integration/
│   └── authentication_flow_test.go
└── performance/
    └── throughput_test.go
```

## Running Tests
```bash
# All tests
make test

# Specific type
make test-unit
make test-integration
make test-performance
```

## Writing Tests
See [DEVELOPER_GUIDE.md](../DEVELOPER_GUIDE.md#testing-strategies)

📍 **Navigation**: [Main](../README.md) | [Developer Guide](../DEVELOPER_GUIDE.md)
