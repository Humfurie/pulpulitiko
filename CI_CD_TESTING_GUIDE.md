# CI/CD and Testing Guide

## Overview

This document describes the comprehensive testing and CI/CD setup for the Pulpulitiko project.

## Table of Contents

- [Test Coverage](#test-coverage)
- [Running Tests Locally](#running-tests-locally)
- [CI/CD Workflows](#cicd-workflows)
- [Docker Testing](#docker-testing)
- [Code Quality Standards](#code-quality-standards)

---

## Test Coverage

### Backend Tests (Go)

**Location**: `/api`

**Test Files** (27 tests total):
- `internal/repository/import_repository_test.go` (8 tests, 570 lines)
- `internal/services/import_service_test.go` (9 tests, 645 lines)
- `internal/handlers/import_handler_test.go` (10 tests, 815 lines)

**Coverage**: >80% target for all packages

**What's Tested**:
- ✅ Import validation logic
- ✅ Async import processing
- ✅ Database operations (CRUD)
- ✅ HTTP handlers and endpoints
- ✅ Template generation
- ✅ Error report generation
- ✅ Edge cases and error handling

### Frontend Tests (Vue/Nuxt)

**Location**: `/web`

**Test Files** (40 tests total):
- `app/composables/useGrouping.test.ts` (40 tests)

**Coverage**: >90% target (100% achieved for useGrouping)

**What's Tested**:
- ✅ Grouping logic
- ✅ Expand/collapse functionality
- ✅ localStorage persistence
- ✅ Reactivity and state management
- ✅ Edge cases (null, undefined, special characters)
- ✅ Performance with large datasets
- ✅ TypeScript types

---

## Running Tests Locally

### Backend (Go)

```bash
cd api

# Run all tests
go test ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run with race detection
go test -race ./...

# Run specific package
go test ./internal/services

# Run specific test
go test -run TestValidateImport ./internal/services

# Verbose output
go test -v ./...
```

### Frontend (Vitest)

```bash
cd web

# Install dependencies (first time)
npm install

# Run all tests
npm test

# Run tests once (CI mode)
npm run test:run

# Run with coverage
npm run test:coverage

# Run with UI
npm run test:ui

# Watch mode
npm test -- --watch

# Run specific test file
npm test -- useGrouping.test.ts
```

### Docker-based Testing

```bash
# Run all tests in Docker
docker compose -f docker-compose.test.yml up --abort-on-container-exit

# View test output
docker compose -f docker-compose.test.yml logs

# Cleanup
docker compose -f docker-compose.test.yml down -v
```

---

## CI/CD Workflows

### 1. Backend CI (`backend-ci.yml`)

**Triggers**:
- Push to `master`, `main`, or `develop`
- Pull requests to these branches
- Changes in `api/**` directory

**Jobs**:

#### Lint and Test
- Go version: 1.24.0
- Services: PostgreSQL 16, Redis 7
- Steps:
  1. Code checkout
  2. Go setup and dependency caching
  3. Code formatting (`go fmt`)
  4. Static analysis (`go vet`, `staticcheck`)
  5. Linting (`golangci-lint`)
  6. Database migrations
  7. Tests with race detection
  8. Coverage reporting (minimum 70%)
  9. Build verification

#### Security Scan
- Gosec security scanner
- Trivy vulnerability scanner
- SARIF results uploaded to GitHub Security

**Environment Variables Required**:
- `DATABASE_URL`: Provided by service container
- `REDIS_URL`: Provided by service container
- `JWT_SECRET`: Set to test value
- `ENVIRONMENT`: Set to "test"

### 2. Frontend CI (`frontend-ci.yml`)

**Triggers**:
- Push to `master`, `main`, or `develop`
- Pull requests to these branches
- Changes in `web/**` directory

**Jobs**:

#### Lint, Test & Build
- Node version: 20
- Steps:
  1. Code checkout
  2. Node setup with npm caching
  3. ESLint
  4. TypeScript type checking
  5. Vitest unit tests
  6. Coverage reporting (90% threshold)
  7. Production build
  8. Build artifact upload

#### Accessibility Test
- axe-core accessibility testing
- Tests key pages for WCAG compliance

#### Security Scan
- npm audit for vulnerabilities
- Snyk security scanning
- Trivy vulnerability scanner

**Environment Variables Required**:
- `NUXT_PUBLIC_API_URL`: Set for build

### 3. Integration Tests (`integration-tests.yml`)

**Triggers**:
- Push to `master`, `main`, or `develop`
- Pull requests
- Manual workflow dispatch

**Jobs**:

#### Docker Integration
- Full stack testing with Docker Compose
- Steps:
  1. Build all Docker images
  2. Start services (postgres, redis, api, web)
  3. Run database migrations
  4. Seed test data
  5. API health checks
  6. Integration tests
  7. Log collection

#### E2E Tests
- Playwright end-to-end tests
- Tests actual user workflows
- Screenshot/video capture on failure

---

## Docker Testing

### Test Environment

The `docker-compose.test.yml` file provides an isolated test environment:

**Services**:
- `postgres-test`: PostgreSQL 16 (port 5433)
- `redis-test`: Redis 7 (port 6380)
- `api-test`: Go backend with test runner
- `web-test`: Nuxt frontend with Vitest

**Network**: Isolated `test-network`

### Running Tests in Docker

```bash
# Run all tests
docker compose -f docker-compose.test.yml up --build --abort-on-container-exit

# View logs
docker compose -f docker-compose.test.yml logs api-test
docker compose -f docker-compose.test.yml logs web-test

# Clean up
docker compose -f docker-compose.test.yml down -v
```

### Integration with CI

The integration tests workflow automatically:
1. Builds Docker images from source
2. Starts all services with health checks
3. Runs migrations and seeds
4. Executes integration tests
5. Collects logs on failure
6. Cleans up resources

---

## Code Quality Standards

### Backend (Go)

**Formatting**:
- Use `go fmt` (enforced in CI)
- No formatting errors allowed

**Linting**:
- `go vet`: Standard Go linting
- `staticcheck`: Advanced static analysis
- `golangci-lint`: Comprehensive linting suite

**Code Coverage**:
- Minimum: 70% overall
- Target: >80% for new code
- Measured with `go test -cover`

**Security**:
- Gosec: Security vulnerability scanning
- Trivy: Dependency vulnerability scanning
- No HIGH or CRITICAL vulnerabilities allowed

### Frontend (Vue/Nuxt)

**Type Safety**:
- TypeScript strict mode
- No type errors allowed (`nuxi typecheck`)

**Linting**:
- ESLint with Nuxt config
- No errors allowed
- Warnings reviewed

**Code Coverage**:
- Minimum: 90% for composables and utilities
- Target: >80% overall
- Thresholds enforced in `vitest.config.ts`

**Security**:
- npm audit: No HIGH vulnerabilities
- Snyk: Continuous dependency monitoring
- Trivy: Container and dependency scanning

**Accessibility**:
- axe-core testing for WCAG 2.1 Level A
- Manual testing for complex interactions

---

## CI/CD Best Practices

### 1. Caching Strategy

**Go**:
```yaml
- uses: actions/cache@v4
  with:
    path: |
      ~/go/pkg/mod
      ~/.cache/go-build
    key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
```

**Node**:
```yaml
- uses: setup-node@v4
  with:
    cache: 'npm'
    cache-dependency-path: web/package-lock.json
```

### 2. Service Containers

Use GitHub Actions services for dependencies:
```yaml
services:
  postgres:
    image: postgres:16-alpine
    options: --health-cmd pg_isready
```

### 3. Parallel Jobs

Independent jobs run in parallel:
- Linting and testing
- Security scanning
- Build verification

### 4. Artifact Management

Build artifacts are uploaded and shared:
```yaml
- uses: actions/upload-artifact@v4
  with:
    name: frontend-build
    path: web/.output
    retention-days: 7
```

### 5. Security Integration

SARIF results uploaded to GitHub Security tab:
```yaml
- uses: github/codeql-action/upload-sarif@v3
  with:
    sarif_file: 'trivy-results.sarif'
```

---

## Troubleshooting

### Backend Tests Failing

**Check**:
1. Database connection: `docker logs pulpulitiko-postgres-test`
2. Environment variables are set correctly
3. Migrations have run: `migrate -version`
4. Mock setup in tests is correct

**Common Issues**:
- Database not ready: Increase health check wait time
- Port conflicts: Ensure ports 5432, 6379 are available
- Race conditions: Run with `-race` flag to detect

### Frontend Tests Failing

**Check**:
1. Dependencies installed: `npm ci`
2. Node version matches CI (20)
3. Environment variables set
4. localStorage mock is working

**Common Issues**:
- Snapshot mismatches: Update snapshots with `-u`
- Async timing: Use `waitFor` helpers
- DOM not available: Check test environment setup

### CI/CD Workflow Failing

**Check**:
1. GitHub Actions logs in detail
2. Service health checks passing
3. Required secrets are set
4. Workflow permissions are correct

**Common Issues**:
- Cache corruption: Clear cache and re-run
- Service timeouts: Increase health check intervals
- Artifact upload limits: Check file sizes

---

## Coverage Reports

### Viewing Coverage Locally

**Backend**:
```bash
cd api
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

**Frontend**:
```bash
cd web
npm run test:coverage
# Open web/coverage/index.html in browser
```

### Coverage in CI

Coverage reports are uploaded to Codecov:
- Backend: `codecov.io/gh/[org]/pulpulitiko?flag=backend`
- Frontend: `codecov.io/gh/[org]/pulpulitiko?flag=frontend`

---

## Adding New Tests

### Backend Test Template

```go
package services_test

import (
    "context"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestNewFeature(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {
            name:    "valid input",
            input:   "test",
            want:    "expected",
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := NewFeature(tt.input)
            if tt.wantErr {
                require.Error(t, err)
                return
            }
            require.NoError(t, err)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

### Frontend Test Template

```typescript
import { describe, it, expect, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'

describe('NewComponent', () => {
  beforeEach(() => {
    // Setup
  })

  it('should render correctly', () => {
    const wrapper = mount(NewComponent)
    expect(wrapper.exists()).toBe(true)
  })

  it('should handle user interaction', async () => {
    const wrapper = mount(NewComponent)
    await wrapper.find('button').trigger('click')
    expect(wrapper.emitted('submit')).toBeTruthy()
  })
})
```

---

## Maintenance

### Updating Dependencies

**Backend**:
```bash
cd api
go get -u ./...
go mod tidy
go test ./...
```

**Frontend**:
```bash
cd web
npm update
npm audit fix
npm test
```

### Updating CI Workflows

1. Test changes locally with Docker
2. Create feature branch
3. Update workflow file
4. Test in PR
5. Monitor first production run
6. Merge when stable

---

## Additional Resources

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Vitest Documentation](https://vitest.dev/)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Docker Compose Testing](https://docs.docker.com/compose/)
- [Codecov Documentation](https://docs.codecov.com/)

---

## Support

For issues with CI/CD or testing:
1. Check this guide
2. Review GitHub Actions logs
3. Check test output locally
4. Review Docker logs
5. Create issue with reproduction steps
