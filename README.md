# Services API

A complete Go-based microservices API built with Gin, MySQL, Docker, and comprehensive CI/CD pipeline.

## 🚀 Quick Start

### Prerequisites

- Docker & Docker Compose
- Go 1.22+ (for local development)
- MySQL 8.4+ (if running locally without Docker)

### Start Everything with Docker

```bash
# Start MySQL + API (builds the image)
make dev
```

This will:
- Build the API Docker image
- Start MySQL with health checks
- Run database migrations
- Start the API server on port 8080

### API Endpoints

Once running, the API will be available at `http://localhost:8080`:

- `GET /health` - Health check
- `GET /swagger/index.html` - **Swagger UI Documentation** 📖
- `GET /api/v1/services` - List all services
- `POST /api/v1/services` - Create a new service
- `GET /api/v1/services/{id}` - Get a specific service
- `PUT /api/v1/services/{id}` - Update a service
- `DELETE /api/v1/services/{id}` - Delete a service
- `GET /api/v1/services/{id}/versions` - List versions for a service
- `POST /api/v1/services/{id}/versions` - Create a new version

### 📖 API Documentation

The API includes comprehensive Swagger/OpenAPI documentation:

- **Interactive Swagger UI**: `http://localhost:8080/swagger/index.html`
- **OpenAPI JSON**: `http://localhost:8080/swagger/doc.json`
- **OpenAPI YAML**: Available in `docs/swagger.yaml`

The Swagger UI provides:
- Complete API reference with request/response examples
- Interactive "Try it out" functionality
- Schema definitions for all models
- Parameter descriptions and validation rules

### Example API Usage

```bash
# Create a service
curl -X POST http://localhost:8080/api/v1/services \
  -H "Content-Type: application/json" \
  -d '{"name":"My Service","slug":"my-service","description":"A test service"}'

# List all services
curl http://localhost:8080/api/v1/services

# Create a version
curl -X POST http://localhost:8080/api/v1/services/{service-id}/versions \
  -H "Content-Type: application/json" \
  -d '{"semver":"1.0.0","status":"released","changelog":"Initial release"}'
```

## 🛠️ Development

### Local Development (without Docker)

1. **Start MySQL:**
   ```bash
   docker compose up -d mysql
   ```

2. **Run migrations:**
   ```bash
   make migrate-up
   ```

3. **Build and run the API:**
   ```bash
   make build
   MYSQL_DSN=app:app@tcp(127.0.0.1:3306)/servicesdb?parseTime=true make run
   ```

### Available Make Commands

- `make dev` - Start everything with Docker Compose
- `make build` - Build the Go binary locally
- `make test` - Run tests
- `make lint` - Run linter
- `make fmt` - Format code
- `make ci` - Run full CI pipeline (format, lint, test, coverage)
- `make docker` - Build Docker image
- `make docker-run` - Run Docker container
- `make docker-push` - Push to registry (requires login)
- `make migrate-up` - Run database migrations
- `make migrate-down` - Rollback migrations
- `make seed` - Load demo data
- `make coverage` - Generate test coverage report
- `make docs` - Generate Swagger documentation
- `make clean` - Clean build artifacts

### Database Management

The project uses [Goose](https://github.com/pressly/goose) for database migrations:

```bash
# Install goose
go install github.com/pressly/goose/v3/cmd/goose@latest

# Run migrations
make migrate-up

# Rollback migrations
make migrate-down

# Load demo data
make seed
```

### Swagger Documentation

The API uses Swagger/OpenAPI for documentation generation:

```bash
# Generate Swagger docs (automatically done during build)
make docs

# Install swag tool manually if needed
go install github.com/swaggo/swag/cmd/swag@v1.16.3
```

Documentation is automatically generated from code annotations and includes:
- Complete API specification
- Request/response schemas
- Interactive testing interface
- Parameter validation rules

## 🏗️ Project Structure

```
.
├── cmd/
│   └── api/
│       └── main.go              # Application entry point
├── internal/                    # Private application code
│   ├── app/                     # Application business logic
│   ├── adapters/                # External adapters
│   │   ├── http/                # HTTP handlers
│   │   └── db/                  # Database repositories
│   ├── domain/                  # Domain models
│   ├── ports/                   # Interface definitions
│   └── config/                  # Configuration
├── migrations/                  # Database migrations
│   ├── 0001_init.sql           # Initial schema
│   └── 0002_demo_seed.sql      # Demo data
├── docs/                       # Generated Swagger documentation
│   ├── docs.go                 # Generated Go docs
│   ├── swagger.json            # OpenAPI JSON spec
│   └── swagger.yaml            # OpenAPI YAML spec
├── build/
│   └── docker/
│       └── Dockerfile          # Multi-stage Docker build
├── .github/
│   └── workflows/
│       └── ci.yml              # GitHub Actions CI/CD
├── .golangci.yml               # Linting configuration
├── docker-compose.yml          # Local development setup
├── Makefile                    # Build automation
└── .env.example               # Environment variables template
```

## 🐳 Docker

### Multi-stage Build

The Dockerfile uses a multi-stage build for optimal image size:

1. **Build stage:** Go 1.22 with build tools and caching
2. **Runtime stage:** Distroless image with just the binary

### Image Registry

Images are pushed to GitHub Container Registry (GHCR):
- `ghcr.io/{owner}/{repo}/services-api:latest`
- `ghcr.io/{owner}/{repo}/services-api:{git-sha}`

To customize the image name, update the `IMAGE` variable in the Makefile.

## 🔄 CI/CD Pipeline

The GitHub Actions workflow (`.github/workflows/ci.yml`) provides:

### On Pull Requests & Main Branch:
- Go setup with caching
- Swagger documentation generation
- Code linting with golangci-lint
- Unit tests with race detection
- Database integration tests
- Code coverage reporting

### On Main Branch (additional):
- Docker image build
- Multi-platform support (linux/amd64)
- Automatic push to GHCR
- Semantic versioning with git SHA

### Required Secrets:
- `GITHUB_TOKEN` - Automatically provided for GHCR access

## 🔧 Configuration

### Environment Variables

Copy `.env.example` to `.env` and customize:

```env
PORT=8080
LOG_LEVEL=info
MYSQL_DSN=app:app@tcp(127.0.0.1:3306)/servicesdb?parseTime=true&charset=utf8mb4&collation=utf8mb4_0900_ai_ci
```

### Database Schema

The API manages two main entities:

- **Services:** Core service definitions with name, slug, description
- **Versions:** Semantic versions for each service with status tracking

Features:
- UUID primary keys
- Automatic timestamps
- Foreign key constraints
- Full-text search on service names/descriptions
- Denormalized version counts with triggers

## 🧪 Testing

```bash
# Run all tests
make test

# Run tests with coverage
make coverage

# Run tests in CI mode (with race detection)
make ci
```

## 🚀 Deployment

### Local Deployment
```bash
make dev
```

### Production Deployment
1. Push to main branch
2. GitHub Actions builds and pushes Docker image
3. Deploy the image from GHCR to your container platform

### Manual Docker Deployment
```bash
# Build image
make docker

# Run container
make docker-run

# Or with custom settings
docker run --rm -p 8080:8080 \
  -e MYSQL_DSN="your-dsn-here" \
  -e LOG_LEVEL=info \
  ghcr.io/your-org/services-api:latest
```

## 📝 API Documentation

The API follows RESTful conventions:

### Service Model
```json
{
  "id": "uuid",
  "name": "Service Name",
  "slug": "service-slug",
  "description": "Service description",
  "created_at": "2023-01-01T00:00:00Z",
  "updated_at": "2023-01-01T00:00:00Z",
  "versions_count": 3
}
```

### Version Model
```json
{
  "id": "uuid",
  "service_id": "uuid",
  "semver": "1.0.0",
  "status": "released",
  "changelog": "Release notes",
  "created_at": "2023-01-01T00:00:00Z"
}
```

### Status Values
- `draft` - Work in progress
- `released` - Available for use
- `deprecated` - No longer recommended

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run `make ci` to ensure quality
5. Submit a pull request

The CI pipeline will automatically test your changes and provide feedback.

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.
