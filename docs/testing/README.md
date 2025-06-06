# Testing Documentation

This directory contains all testing-related documentation and reports for the Hexabase KaaS project.

## Directory Structure

```
testing/
├── README.md              # This file
├── testing-guide.md       # General testing guide
├── oauth-testing.md       # OAuth testing guide
└── coverage-reports/      # Test coverage and result reports
    ├── test-results.md
    ├── test_coverage_report.md
    └── COVERAGE_REPORT.md
```

## Test Organization

### API Tests (Go)
- **Unit Tests**: Located alongside source files (e.g., `auth/jwt_test.go`)
- **Integration Tests**: Located in `/api/tests/integration/`
- **E2E Tests**: Located in `/api/tests/e2e/`

### UI Tests (TypeScript)
- **Playwright Tests**: Located in `/ui/tests/`

### Test Scripts
- Located in `/scripts/test/`
- Includes database setup, token generation, and test runners

## Running Tests

### Go Tests
```bash
cd api
go test ./...                    # Run all tests
go test ./... -cover            # Run with coverage
go test -v ./internal/auth/...  # Run specific package tests
```

### UI Tests
```bash
cd ui
npm test                        # Run unit tests
npm run test:e2e               # Run Playwright tests
```

## Test Coverage

Latest test coverage reports are available in the `coverage-reports` directory. These reports are generated automatically during CI/CD runs and can also be generated locally.

### Generating Coverage Reports

For Go:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

For TypeScript:
```bash
npm run test:coverage
```