# Testing Documentation

This directory contains guides and best practices for testing the Hexabase AI platform. 

**Important**: All test results and coverage reports have been moved to `/api/testresults/` to maintain a clean project structure.

## Directory Structure

```
testing/
├── README.md              # This file
├── testing-guide.md       # General testing guide and best practices
└── oauth-testing.md       # OAuth integration testing guide
```

## Test Result Locations

Test execution results are now centrally located:

- **Test Results**: `/api/testresults/`
- **Coverage Reports**: `/api/testresults/coverage/`
- **Test Summaries**: `/api/testresults/summary/`
- **Comprehensive Report**: `/api/testresults/COMPREHENSIVE_TEST_REPORT.md`
- **Historical Reports**: `/api/testresults/reports/`

## Test Organization

### API Tests (Go)
- **Unit Tests**: Located alongside source files (e.g., `auth/jwt_test.go`)
- **Integration Tests**: Located in `/api/tests/integration/`
- **E2E Tests**: Located in `/api/tests/e2e/`
- **Test Results**: All results in `/api/testresults/`

### UI Tests (TypeScript)
- **Playwright Tests**: Located in `/ui/tests/`
- **Unit Tests**: Located alongside components
- **Screenshots**: Located in `/ui/screenshots/`

### Test Scripts
- Located in `/scripts/test/`
- Includes database setup, token generation, and test runners

## Running Tests

### Go Tests
```bash
cd api
# Run all tests with coverage and reporting
./run_tests_with_coverage.sh

# Run all tests
go test ./...

# Run with coverage
go test ./... -cover

# Run specific package tests
go test -v ./internal/auth/...

# Run with race detection
go test -race ./...
```

### UI Tests
```bash
cd ui
# Run unit tests
npm test

# Run Playwright tests
npm run test:e2e

# Run specific test file
npm test -- workspace.spec.ts

# Run in headed mode for debugging
npm test -- --headed
```

## Testing Best Practices

### Unit Testing
- Write tests alongside implementation code
- Aim for >80% coverage on critical paths
- Use table-driven tests for multiple scenarios
- Mock external dependencies

### Integration Testing
- Test API endpoints with real HTTP requests
- Use test database for data persistence tests
- Verify error handling and edge cases
- Test authentication and authorization

### End-to-End Testing
- Use Playwright for UI testing
- Test complete user workflows
- Capture screenshots for visual validation
- Run against staging environment

## Coverage Requirements

- **Critical Paths**: >80% coverage required
- **Service Layer**: >70% coverage required
- **Handlers**: >60% coverage required
- **Utilities**: >50% coverage required

## Generating Coverage Reports

The test script automatically generates comprehensive reports:

```bash
cd api
./run_tests_with_coverage.sh
```

This will create:
- Coverage reports in `/api/testresults/coverage/[timestamp]/`
- Test summaries in `/api/testresults/summary/`
- HTML coverage visualization
- Package-level coverage details

For manual coverage generation:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## Additional Resources

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Playwright Documentation](https://playwright.dev/)
- [Test Coverage in Go](https://blog.golang.org/cover)
- [Table Driven Tests](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)

---

For actual test results and coverage reports, see `/api/testresults/`.