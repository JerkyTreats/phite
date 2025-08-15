# github.com/JerkyTreats/PHITE/vital-hearth/backend

A Go API template with automated testing, OpenAPI documentation generation, and Docker deployment.

## Features

- **Route Registry System**: Automatic route discovery and registration
- **OpenAPI Documentation**: Auto-generated API documentation with zero maintenance
- **Swagger UI & ReDoc**: Interactive API documentation interfaces
- **GitHub Actions CI/CD**: Automated testing, linting, and Docker image building
- **Structured Logging**: Configurable logging with Zap
- **Configuration Management**: Flexible config with Viper
- **Health Checks**: Built-in health endpoints
- **Docker Support**: Multi-stage builds with security best practices

## Quick Start

1. **Initialize the template**:
   ```bash
   ./scripts/template-init.sh
   ```

2. **Install dependencies**:
   ```bash
   go mod tidy
   ```

3. **Run the application**:
   ```bash
   go run cmd/server/main.go
   ```

4. **Generate OpenAPI documentation**:
   ```bash
   go run cmd/generate-openapi/main.go
   ```

5. **Access API documentation**:
   - Swagger UI: http://localhost:8080/docs
   - ReDoc: http://localhost:8080/redoc
   - OpenAPI JSON: http://localhost:8080/api/docs/openapi.json
   - OpenAPI YAML: http://localhost:8080/api/docs/openapi.yaml

## Project Structure

```
├── .github/workflows/    # GitHub Actions CI/CD
├── cmd/
│   ├── server/          # Main application
│   └── generate-openapi/ # OpenAPI spec generator
├── internal/
│   ├── api/             # API handlers and types
│   ├── config/          # Configuration management
│   └── logging/         # Logging setup
├── docs/api/            # Generated OpenAPI documentation
├── configs/             # Configuration files
└── scripts/             # Utility scripts
```

## Adding New API Endpoints

1. **Create a handler** in `internal/api/handler/`
2. **Register the route** using the route registry:
   ```go
   func init() {
       types.RegisterRoute(types.RouteInfo{
           Method:       "GET",
           Path:         "/my-endpoint",
           Handler:      myHandler,
           ResponseType: reflect.TypeOf(MyResponse{}),
           Module:       "my-module",
           Summary:      "Description of my endpoint",
       })
   }
   ```
3. **Run OpenAPI generation** to update documentation automatically

## Configuration

The application uses Viper for configuration management. Configuration files should be placed in the `configs/` directory.

## Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Docker

```bash
# Build image
docker build -t github.com/JerkyTreats/PHITE/vital-hearth/backend .

# Run container
docker run -p 8080:8080 github.com/JerkyTreats/PHITE/vital-hearth/backend
```

## CI/CD

The project includes GitHub Actions workflows for:
- Running tests and linting
- Generating OpenAPI documentation
- Building and pushing Docker images
- Multi-platform builds (linux/amd64, linux/arm64)

Set up the following GitHub secrets:
- `DOCKER_USERNAME`: Your Docker Hub username
- `DOCKER_PASSWORD`: Your Docker Hub password

## OpenAPI Documentation

The project automatically generates OpenAPI 3.0 specifications from your route definitions. The generated documentation includes:
- Request/response schemas derived from Go types
- Automatic operation IDs and descriptions
- Standard HTTP status codes and error responses
- Interactive Swagger UI and ReDoc interfaces

Documentation is generated at `docs/api/openapi.yaml` and `docs/api/swagger.json` and automatically updated by CI/CD.

## Template Initialization

See [TEMPLATE_PLACEHOLDERS.md](TEMPLATE_PLACEHOLDERS.md) for details on template placeholders and initialization.
