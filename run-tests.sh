#!/bin/bash

# Pulpulitiko Test Runner
# This script provides easy commands to run tests locally

set -e

COLOR_GREEN='\033[0;32m'
COLOR_BLUE='\033[0;34m'
COLOR_YELLOW='\033[1;33m'
COLOR_RED='\033[0;31m'
COLOR_RESET='\033[0m'

function print_header() {
    echo -e "\n${COLOR_BLUE}========================================${COLOR_RESET}"
    echo -e "${COLOR_BLUE}  $1${COLOR_RESET}"
    echo -e "${COLOR_BLUE}========================================${COLOR_RESET}\n"
}

function print_success() {
    echo -e "${COLOR_GREEN}✓ $1${COLOR_RESET}"
}

function print_error() {
    echo -e "${COLOR_RED}✗ $1${COLOR_RESET}"
}

function print_info() {
    echo -e "${COLOR_YELLOW}ℹ $1${COLOR_RESET}"
}

function backend_tests() {
    print_header "Running Backend Tests"

    cd api

    print_info "Installing dependencies..."
    go mod download

    print_info "Running tests..."
    if go test -v -race ./...; then
        print_success "Backend tests passed!"
    else
        print_error "Backend tests failed!"
        cd ..
        exit 1
    fi

    cd ..
}

function backend_coverage() {
    print_header "Running Backend Tests with Coverage"

    cd api

    print_info "Running tests with coverage..."
    go test -coverprofile=coverage.out ./...

    print_info "Coverage summary:"
    go tool cover -func=coverage.out | tail -n 1

    print_success "Opening coverage report in browser..."
    go tool cover -html=coverage.out

    cd ..
}

function frontend_tests() {
    print_header "Running Frontend Tests"

    cd web

    if [ ! -d "node_modules" ]; then
        print_info "Installing dependencies..."
        npm ci
    fi

    print_info "Running tests..."
    if npm run test:run; then
        print_success "Frontend tests passed!"
    else
        print_error "Frontend tests failed!"
        cd ..
        exit 1
    fi

    cd ..
}

function frontend_coverage() {
    print_header "Running Frontend Tests with Coverage"

    cd web

    if [ ! -d "node_modules" ]; then
        print_info "Installing dependencies..."
        npm ci
    fi

    print_info "Running tests with coverage..."
    npm run test:coverage

    print_success "Coverage report generated at web/coverage/index.html"

    # Try to open coverage report
    if command -v xdg-open &> /dev/null; then
        xdg-open coverage/index.html
    elif command -v open &> /dev/null; then
        open coverage/index.html
    fi

    cd ..
}

function docker_tests() {
    print_header "Running Tests in Docker"

    print_info "Building and running test containers..."
    if docker compose -f docker-compose.test.yml up --build --abort-on-container-exit; then
        print_success "Docker tests completed!"
    else
        print_error "Docker tests failed!"
        docker compose -f docker-compose.test.yml down -v
        exit 1
    fi

    print_info "Cleaning up..."
    docker compose -f docker-compose.test.yml down -v
}

function lint_backend() {
    print_header "Running Backend Linters"

    cd api

    print_info "Running go fmt..."
    if [ -n "$(gofmt -l .)" ]; then
        print_error "Code is not formatted. Run: cd api && gofmt -w ."
        gofmt -l .
        cd ..
        exit 1
    fi
    print_success "go fmt check passed"

    print_info "Running go vet..."
    go vet ./...
    print_success "go vet passed"

    if command -v staticcheck &> /dev/null; then
        print_info "Running staticcheck..."
        staticcheck ./...
        print_success "staticcheck passed"
    else
        print_info "staticcheck not installed, skipping..."
    fi

    if command -v golangci-lint &> /dev/null; then
        print_info "Running golangci-lint..."
        golangci-lint run
        print_success "golangci-lint passed"
    else
        print_info "golangci-lint not installed, skipping..."
    fi

    cd ..
}

function lint_frontend() {
    print_header "Running Frontend Linters"

    cd web

    if [ ! -d "node_modules" ]; then
        print_info "Installing dependencies..."
        npm ci
    fi

    print_info "Running ESLint..."
    npm run lint
    print_success "ESLint passed"

    print_info "Running TypeScript type check..."
    npm run typecheck
    print_success "Type check passed"

    cd ..
}

function all_tests() {
    print_header "Running All Tests"

    lint_backend
    backend_tests
    lint_frontend
    frontend_tests

    print_success "All tests passed! 🎉"
}

function quick_test() {
    print_header "Quick Test (No Linting)"

    backend_tests
    frontend_tests

    print_success "Quick tests passed! 🎉"
}

function show_help() {
    cat << EOF
Pulpulitiko Test Runner

Usage: ./run-tests.sh [command]

Commands:
  backend           Run backend tests
  backend-coverage  Run backend tests with coverage report
  frontend          Run frontend tests
  frontend-coverage Run frontend tests with coverage report
  docker            Run all tests in Docker
  lint-backend      Run backend linters only
  lint-frontend     Run frontend linters only
  all               Run all tests and linters (default)
  quick             Run tests without linting
  help              Show this help message

Examples:
  ./run-tests.sh                    # Run all tests
  ./run-tests.sh backend            # Run only backend tests
  ./run-tests.sh frontend-coverage  # Frontend tests with coverage
  ./run-tests.sh docker             # Run in Docker

For more information, see:
  - CI_CD_TESTING_GUIDE.md
  - TESTING_SUMMARY.md
  - api/internal/TEST_README.md
  - web/test/README.md
EOF
}

# Main script logic
case "${1:-all}" in
    backend)
        backend_tests
        ;;
    backend-coverage)
        backend_coverage
        ;;
    frontend)
        frontend_tests
        ;;
    frontend-coverage)
        frontend_coverage
        ;;
    docker)
        docker_tests
        ;;
    lint-backend)
        lint_backend
        ;;
    lint-frontend)
        lint_frontend
        ;;
    all)
        all_tests
        ;;
    quick)
        quick_test
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        print_error "Unknown command: $1"
        echo ""
        show_help
        exit 1
        ;;
esac
