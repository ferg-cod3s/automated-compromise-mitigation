#!/usr/bin/env bash
# Test runner with coverage reporting for ACM
# Targets >80% code coverage for security-critical application

set -euo pipefail

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"

# Configuration
COVERAGE_THRESHOLD=${COVERAGE_THRESHOLD:-80}
COVERAGE_FILE="${ROOT_DIR}/coverage.out"
COVERAGE_HTML="${ROOT_DIR}/coverage.html"
COVERAGE_JSON="${ROOT_DIR}/coverage.json"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $*"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $*"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $*" >&2
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $*"
}

# Check prerequisites
check_prerequisites() {
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed"
        exit 1
    fi

    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    log_info "Go version: $GO_VERSION"
}

# Clean previous coverage data
clean_coverage() {
    log_info "Cleaning previous coverage data..."
    rm -f "$COVERAGE_FILE" "$COVERAGE_HTML" "$COVERAGE_JSON"
}

# Run tests with coverage
run_tests() {
    log_info "Running tests with coverage..."
    echo ""

    cd "$ROOT_DIR"

    # Run tests with coverage, race detection, and verbose output
    go test -v \
        -race \
        -timeout 10m \
        -coverprofile="$COVERAGE_FILE" \
        -covermode=atomic \
        -coverpkg=./... \
        ./... 2>&1 | tee test-output.log

    local TEST_EXIT_CODE=${PIPESTATUS[0]}

    echo ""

    if [ $TEST_EXIT_CODE -ne 0 ]; then
        log_error "Tests failed with exit code $TEST_EXIT_CODE"
        return $TEST_EXIT_CODE
    fi

    log_success "All tests passed"
    return 0
}

# Generate coverage report
generate_coverage_report() {
    if [ ! -f "$COVERAGE_FILE" ]; then
        log_error "Coverage file not found: $COVERAGE_FILE"
        return 1
    fi

    log_info "Generating coverage reports..."

    # Generate HTML coverage report
    go tool cover -html="$COVERAGE_FILE" -o "$COVERAGE_HTML"
    log_success "HTML coverage report: $COVERAGE_HTML"

    # Generate text coverage report
    go tool cover -func="$COVERAGE_FILE" > coverage-report.txt
    log_success "Text coverage report: coverage-report.txt"
}

# Calculate and display coverage
calculate_coverage() {
    if [ ! -f "$COVERAGE_FILE" ]; then
        log_error "Coverage file not found: $COVERAGE_FILE"
        return 1
    fi

    log_info "Calculating coverage..."
    echo ""

    # Calculate total coverage
    local TOTAL_COVERAGE=$(go tool cover -func="$COVERAGE_FILE" | grep total | awk '{print $3}' | sed 's/%//')

    echo "╔════════════════════════════════════════╗"
    echo "║        COVERAGE REPORT                 ║"
    echo "╠════════════════════════════════════════╣"
    printf "║  Total Coverage:  %5.1f%%              ║\n" "$TOTAL_COVERAGE"
    printf "║  Target:          %5.1f%%              ║\n" "$COVERAGE_THRESHOLD"
    echo "╚════════════════════════════════════════╝"
    echo ""

    # Check if coverage meets threshold
    if (( $(echo "$TOTAL_COVERAGE >= $COVERAGE_THRESHOLD" | bc -l) )); then
        log_success "Coverage threshold met! (${TOTAL_COVERAGE}% >= ${COVERAGE_THRESHOLD}%)"
        return 0
    else
        log_warn "Coverage below threshold! (${TOTAL_COVERAGE}% < ${COVERAGE_THRESHOLD}%)"
        return 1
    fi
}

# Show coverage by package
show_package_coverage() {
    if [ ! -f "$COVERAGE_FILE" ]; then
        return 1
    fi

    log_info "Coverage by package:"
    echo ""

    # Extract package-level coverage
    go tool cover -func="$COVERAGE_FILE" | grep -v "total:" | \
        awk '{package=$1; sub(/\/[^\/]+$/, "", package); cov+=$3; count++}
             END {if(count>0) print package, cov/count"%"}' | \
        sort -k2 -rn | \
        while read -r package coverage; do
            printf "  %-50s %s\n" "$package" "$coverage"
        done

    echo ""
}

# Find untested packages/files
find_untested() {
    if [ ! -f "$COVERAGE_FILE" ]; then
        return 1
    fi

    log_info "Files with low coverage (<50%):"
    echo ""

    go tool cover -func="$COVERAGE_FILE" | \
        awk 'NR>1 && $NF!="total:" {
            coverage = $NF;
            sub(/%/, "", coverage);
            if (coverage+0 < 50) print $1, coverage"%"
        }' | \
        while read -r file coverage; do
            printf "  ${YELLOW}%-60s %s${NC}\n" "$file" "$coverage"
        done

    echo ""
}

# Run integration tests
run_integration_tests() {
    log_info "Running integration tests..."
    echo ""

    cd "$ROOT_DIR"

    if [ ! -d "tests/integration" ]; then
        log_warn "No integration tests found (tests/integration directory missing)"
        return 0
    fi

    go test -v \
        -race \
        -timeout 15m \
        -tags=integration \
        ./tests/integration/...

    local EXIT_CODE=$?

    echo ""

    if [ $EXIT_CODE -ne 0 ]; then
        log_error "Integration tests failed"
        return $EXIT_CODE
    fi

    log_success "Integration tests passed"
    return 0
}

# Run security tests
run_security_tests() {
    log_info "Running security tests..."
    echo ""

    cd "$ROOT_DIR"

    if [ ! -d "tests/security" ]; then
        log_warn "No security tests found (tests/security directory missing)"
        return 0
    fi

    go test -v \
        -race \
        -timeout 15m \
        -tags=security \
        ./tests/security/...

    local EXIT_CODE=$?

    echo ""

    if [ $EXIT_CODE -ne 0 ]; then
        log_error "Security tests failed"
        return $EXIT_CODE
    fi

    log_success "Security tests passed"
    return 0
}

# Generate test summary
generate_summary() {
    log_info "Test Summary"
    echo ""

    if [ -f "test-output.log" ]; then
        local TOTAL_TESTS=$(grep -c "^=== RUN" test-output.log || echo "0")
        local PASSED_TESTS=$(grep -c "^--- PASS:" test-output.log || echo "0")
        local FAILED_TESTS=$(grep -c "^--- FAIL:" test-output.log || echo "0")

        echo "Total Tests:  $TOTAL_TESTS"
        echo "Passed:       $PASSED_TESTS"
        echo "Failed:       $FAILED_TESTS"
        echo ""
    fi

    if [ -f "$COVERAGE_FILE" ]; then
        local TOTAL_COVERAGE=$(go tool cover -func="$COVERAGE_FILE" | grep total | awk '{print $3}')
        echo "Coverage:     $TOTAL_COVERAGE"
        echo ""
    fi

    echo "Reports generated:"
    [ -f "$COVERAGE_FILE" ] && echo "  - $COVERAGE_FILE"
    [ -f "$COVERAGE_HTML" ] && echo "  - $COVERAGE_HTML"
    [ -f "coverage-report.txt" ] && echo "  - coverage-report.txt"
    [ -f "test-output.log" ] && echo "  - test-output.log"
    echo ""
}

# Main execution
main() {
    local RUN_INTEGRATION=${RUN_INTEGRATION:-false}
    local RUN_SECURITY=${RUN_SECURITY:-false}
    local EXIT_CODE=0

    log_info "ACM Test Runner with Coverage"
    echo ""

    check_prerequisites
    clean_coverage

    # Run unit tests
    if ! run_tests; then
        EXIT_CODE=1
    fi

    # Generate coverage reports
    if [ -f "$COVERAGE_FILE" ]; then
        generate_coverage_report
        show_package_coverage
        find_untested

        # Check coverage threshold (warning only, don't fail)
        if ! calculate_coverage; then
            log_warn "Consider adding more tests to improve coverage"
        fi
    fi

    # Run integration tests if requested
    if [ "$RUN_INTEGRATION" = "true" ]; then
        if ! run_integration_tests; then
            EXIT_CODE=1
        fi
    fi

    # Run security tests if requested
    if [ "$RUN_SECURITY" = "true" ]; then
        if ! run_security_tests; then
            EXIT_CODE=1
        fi
    fi

    # Generate summary
    generate_summary

    if [ $EXIT_CODE -eq 0 ]; then
        log_success "All tests completed successfully!"
    else
        log_error "Some tests failed!"
    fi

    return $EXIT_CODE
}

# Run main function
main "$@"
