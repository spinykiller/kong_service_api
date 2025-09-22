# Testing Guide

This document describes the comprehensive testing strategy for the Kong Service API.

## 🧪 Test Structure

### **Unit Tests** (`main_test.go`)
- **Purpose**: Test individual functions and components in isolation
- **Coverage**: Helper functions, data structures, business logic
- **Dependencies**: No external dependencies (database, network)

### **Integration Tests** (`integration_test.go`)
- **Purpose**: Test complete API endpoints with real database
- **Coverage**: Full request/response cycles, database interactions
- **Dependencies**: MySQL database, HTTP requests

## 🚀 Running Tests

### **All Tests**
```bash
make test
```

### **Unit Tests Only**
```bash
make test-unit
```

### **Integration Tests Only**
```bash
# Using local MySQL (requires MySQL running on port 3306)
make test-integration

# Using Docker MySQL (automatically starts/stops test database)
make test-integration-docker
```

### **Test Coverage**
```bash
# Generate coverage report
make test-coverage

# View coverage in browser
open coverage.html
```

## 📋 Test Categories

### **Unit Tests**

#### **Health Check Tests**
- `TestHealthCheck`: Verifies health endpoint returns correct response

#### **Pagination Tests**
- `TestGetPaginationParams`: Tests parameter parsing with defaults
- `TestCalculatePagination`: Tests pagination metadata calculation

#### **Search Tests**
- `TestGetSearchParams`: Tests search parameter parsing

#### **Data Structure Tests**
- `TestServiceStruct`: JSON marshaling/unmarshaling
- `TestVersionStruct`: JSON marshaling/unmarshaling
- `TestPaginatedResponseStruct`: Response structure validation

### **Integration Tests**

#### **API Endpoint Tests**
- `TestHealthCheckIntegration`: Health endpoint with real server
- `TestGetServicesIntegration`: Services listing with pagination
- `TestSearchServicesIntegration`: Full-text search functionality
- `TestCreateServiceIntegration`: Service creation with validation
- `TestGetServiceIntegration`: Individual service retrieval
- `TestGetVersionsIntegration`: Version listing with pagination
- `TestCreateVersionIntegration`: Version creation with transactions

## 🔧 Test Configuration

### **Environment Variables**
- `TEST_MYSQL_DSN`: Test database connection string
- Default: `app:app@tcp(localhost:3306)/servicesdb_test?parseTime=true&charset=utf8mb4&collation=utf8mb4_0900_ai_ci`

### **Test Database Setup**
- **Database**: `servicesdb_test`
- **Tables**: `services`, `versions` (same schema as production)
- **Data**: Seeded with test data for consistent testing
- **Cleanup**: Automatic cleanup after integration tests

## 📊 Test Data

### **Test Services**
```sql
INSERT INTO services (id, name, slug, description) VALUES
  ('service-1', 'Test Service 1', 'test-service-1', 'First test service'),
  ('service-2', 'Test Service 2', 'test-service-2', 'Second test service'),
  ('service-3', 'Notification Service', 'notification-service', 'Service for sending notifications');
```

### **Test Versions**
```sql
INSERT INTO versions (id, service_id, semver, status, changelog) VALUES
  ('version-1', 'service-1', '1.0.0', 'released', 'Initial release'),
  ('version-2', 'service-1', '1.1.0', 'released', 'Minor update'),
  ('version-3', 'service-2', '0.1.0', 'draft', 'Work in progress'),
  ('version-4', 'service-3', '2.0.0', 'released', 'Major update');
```

## 🎯 Test Scenarios

### **Pagination Testing**
- ✅ Default pagination (page=1, page_size=10)
- ✅ Custom pagination parameters
- ✅ Edge cases (empty results, single page)
- ✅ Invalid parameters (page_size > 100, page < 1)

### **Search Testing**
- ✅ Full-text search functionality
- ✅ Relevance-based ordering
- ✅ Pagination with search results
- ✅ Empty search results
- ✅ Missing search query validation

### **CRUD Operations**
- ✅ Create services with validation
- ✅ Duplicate name/slug handling
- ✅ Retrieve services by ID
- ✅ Update services
- ✅ Delete services
- ✅ Version creation with transactions
- ✅ Foreign key constraints

### **Error Handling**
- ✅ Invalid input validation
- ✅ Database error handling
- ✅ HTTP status code verification
- ✅ Error message consistency

## 🔍 Coverage Analysis

### **Coverage Targets**
- **Functions**: > 80%
- **Lines**: > 75%
- **Branches**: > 70%

### **Coverage Commands**
```bash
# Generate coverage report
make test-coverage

# View detailed coverage
go tool cover -func=coverage.out

# Open HTML coverage report
open coverage.html
```

## 🐳 Docker Testing

### **Test Database Container**
```yaml
# docker-compose.test.yml
services:
  mysql-test:
    image: mysql:8.4
    environment:
      MYSQL_DATABASE: servicesdb_test
      MYSQL_USER: app
      MYSQL_PASSWORD: app
    ports:
      - "3307:3306"
```

### **Benefits**
- **Isolation**: No interference with local MySQL
- **Consistency**: Same environment across developers
- **Automation**: Automatic setup/teardown
- **CI/CD Ready**: Easy integration with build pipelines

## 🚀 CI/CD Integration

### **GitHub Actions Example**
```yaml
- name: Run Tests
  run: |
    make test-integration-docker
    make test-coverage
```

### **Test Commands for CI**
```bash
# Full test suite
make test

# With coverage
make test-coverage

# Docker-based integration tests
make test-integration-docker
```

## 📝 Best Practices

### **Test Organization**
- **Unit tests**: Fast, isolated, no external dependencies
- **Integration tests**: Complete workflows, real database
- **Test data**: Consistent, predictable, minimal

### **Naming Conventions**
- **Unit tests**: `TestFunctionName`
- **Integration tests**: `TestFeatureIntegration`
- **Test functions**: Descriptive names explaining the scenario

### **Assertions**
- **Use testify**: `assert.Equal`, `require.NoError`
- **Specific assertions**: Test exact values, not just "not nil"
- **Error handling**: Always test error cases

### **Test Data**
- **Consistent**: Same data across test runs
- **Minimal**: Only what's needed for the test
- **Cleanup**: Always clean up after tests

## 🔧 Troubleshooting

### **Common Issues**

#### **Database Connection Errors**
```bash
# Check if MySQL is running
mysql --protocol tcp -u app -papp -h 127.0.0.1 -P 3306 -e "SELECT 1"

# Check test database exists
mysql --protocol tcp -u app -papp -h 127.0.0.1 -P 3306 -e "SHOW DATABASES"
```

#### **Port Conflicts**
```bash
# Use Docker test database (port 3307)
make test-integration-docker
```

#### **Permission Issues**
```bash
# Ensure MySQL user has proper permissions
mysql --protocol tcp -u root -proot -h 127.0.0.1 -P 3306 -e "GRANT ALL PRIVILEGES ON servicesdb_test.* TO 'app'@'%';"
```

### **Debug Mode**
```bash
# Run tests with verbose output
go test -v ./cmd/api

# Run specific test
go test -run TestGetServicesIntegration ./cmd/api
```

## 📈 Performance Testing

### **Load Testing** (Future Enhancement)
- **Concurrent requests**: Test pagination under load
- **Database performance**: Query optimization
- **Memory usage**: Monitor resource consumption

### **Benchmark Tests** (Future Enhancement)
```go
func BenchmarkGetServices(b *testing.B) {
    // Benchmark service retrieval
}
```

This comprehensive testing strategy ensures the API is robust, reliable, and ready for production use!
