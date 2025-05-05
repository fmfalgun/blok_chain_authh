#!/bin/bash

# Run All Tests Script
# This script runs all tests for the blockchain authentication framework

# Define color codes for output
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Function for printing colored output
print_green() {
    echo -e "${GREEN}$1${NC}"
}

print_yellow() {
    echo -e "${YELLOW}$1${NC}"
}

print_red() {
    echo -e "${RED}$1${NC}"
}

print_yellow "Running All Tests for Blockchain Authentication Framework"
echo "==============================================================="

# Create test results directory
TEST_RESULTS_DIR="test-results-$(date +%Y%m%d%H%M%S)"
mkdir -p $TEST_RESULTS_DIR

print_yellow "Test results will be saved in: $TEST_RESULTS_DIR"
echo "---------------------------------------------------------------"

# Run setup script
print_yellow "Setting up test environment..."
chmod +x setup-test-environment.sh
./setup-test-environment.sh

if [ $? -ne 0 ]; then
    print_red "Test environment setup failed. Aborting tests."
    exit 1
fi

echo "---------------------------------------------------------------"

# Function to run a test and save results
run_test() {
    local test_script=$1
    local test_name=$2
    
    print_yellow "Running $test_name..."
    
    # Make sure the script is executable
    chmod +x $test_script
    
    # Create log file name
    local log_file="$TEST_RESULTS_DIR/$(basename $test_script .sh)-results.log"
    
    # Run the test and capture output and return code
    $test_script > $log_file 2>&1
    local test_result=$?
    
    # Check result
    if [ $test_result -eq 0 ]; then
        print_green "✓ $test_name PASSED"
        echo "PASS" > "$TEST_RESULTS_DIR/$(basename $test_script .sh)-status.txt"
    else
        print_red "✗ $test_name FAILED"
        echo "FAIL" > "$TEST_RESULTS_DIR/$(basename $test_script .sh)-status.txt"
    fi
    
    echo "  Log saved to: $log_file"
    echo "---------------------------------------------------------------"
    
    return $test_result
}

# Initialize counters
TOTAL_TESTS=0
PASSED_TESTS=0

# Run test-rsa-keys.sh
TOTAL_TESTS=$((TOTAL_TESTS + 1))
run_test "./test-rsa-keys.sh" "RSA Key Tests"
if [ $? -eq 0 ]; then
    PASSED_TESTS=$((PASSED_TESTS + 1))
fi

# Run test-authentication-flow.sh
TOTAL_TESTS=$((TOTAL_TESTS + 1))
run_test "./test-authentication-flow.sh" "Authentication Flow Tests"
if [ $? -eq 0 ]; then
    PASSED_TESTS=$((PASSED_TESTS + 1))
fi

# Create summary report
print_yellow "Creating test summary report..."

cat > "$TEST_RESULTS_DIR/test-summary.md" << EOF
# Blockchain Authentication Framework Test Summary

## Test Run Information
- Date: $(date)
- Total Tests: $TOTAL_TESTS
- Passed Tests: $PASSED_TESTS
- Failed Tests: $((TOTAL_TESTS - PASSED_TESTS))
- Success Rate: $(( PASSED_TESTS * 100 / TOTAL_TESTS ))%

## Test Results

| Test | Status | Log File |
|------|--------|----------|
EOF

# Add each test result to the summary
for test in "$TEST_RESULTS_DIR"/*-status.txt; do
    test_name=$(basename $test -status.txt)
    test_status=$(cat $test)
    log_file="$test_name-results.log"
    
    echo "| $test_name | $test_status | [$log_file](./$log_file) |" >> "$TEST_RESULTS_DIR/test-summary.md"
done

# Add system information
cat >> "$TEST_RESULTS_DIR/test-summary.md" << EOF

## System Information

### Network Status
\`\`\`
$(docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep -E "peer|orderer|ca")
\`\`\`

### Chaincode Versions
\`\`\`
$(docker exec cli peer chaincode list --installed 2>/dev/null || echo "Unable to retrieve chaincode information")
\`\`\`

### Environment Details
- Node.js Version: $(node -v 2>/dev/null || echo "Not available")
- Docker Version: $(docker -v 2>/dev/null || echo "Not available")
- Operating System: $(uname -a 2>/dev/null || echo "Not available")
EOF

# Display summary
if [ $PASSED_TESTS -eq $TOTAL_TESTS ]; then
    print_green "All tests passed! ($PASSED_TESTS/$TOTAL_TESTS)"
else
    print_red "Some tests failed. Passed: $PASSED_TESTS/$TOTAL_TESTS"
fi

print_yellow "Test summary saved to: $TEST_RESULTS_DIR/test-summary.md"
echo "---------------------------------------------------------------"

print_yellow "You can view detailed test logs in the $TEST_RESULTS_DIR directory"
