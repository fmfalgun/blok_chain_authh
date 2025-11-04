#!/bin/bash

TEST_TYPE=$1

if [ -z "$TEST_TYPE" ]; then
  echo "Usage: ./run-tests.sh <test-type>"
  echo "Test types: unit, integration, performance, all"
  exit 1
fi

echo "=========================================="
echo "Running $TEST_TYPE tests..."
echo "=========================================="

case $TEST_TYPE in
  unit)
    echo "Running unit tests for AS chaincode..."
    cd ../chaincodes/as-chaincode && go test -v ./...

    echo "Running unit tests for TGS chaincode..."
    cd ../tgs-chaincode && go test -v ./...

    echo "Running unit tests for ISV chaincode..."
    cd ../isv-chaincode && go test -v ./...
    ;;

  integration)
    echo "Running integration tests..."
    cd integration && go test -v ./...
    ;;

  performance)
    echo "Running performance tests..."
    cd performance && go test -v -bench=. ./...
    ;;

  all)
    echo "Running all tests..."
    $0 unit
    $0 integration
    $0 performance
    ;;

  *)
    echo "Unknown test type: $TEST_TYPE"
    exit 1
    ;;
esac

echo "=========================================="
echo "Tests completed!"
echo "=========================================="
