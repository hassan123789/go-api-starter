# API Documentation

This directory contains the OpenAPI/Swagger specification for the Go API Starter.

## Files

- `openapi.yaml` - OpenAPI 3.1 specification

## Usage

### View Documentation

You can use various tools to visualize the API documentation:

#### Swagger UI (Docker)
```bash
docker run -p 8081:8080 -e SWAGGER_JSON=/api/openapi.yaml -v $(pwd)/api:/api swaggerapi/swagger-ui
```

Then open http://localhost:8081

#### Redoc (Docker)
```bash
docker run -p 8081:80 -e SPEC_URL=/api/openapi.yaml -v $(pwd)/api:/usr/share/nginx/html/api redocly/redoc
```

#### Online Editor
- [Swagger Editor](https://editor.swagger.io/) - Paste the YAML content
- [Stoplight Studio](https://stoplight.io/studio/) - Import the file

### Code Generation

Generate client SDKs or server stubs:

```bash
# Generate Go client
openapi-generator generate -i api/openapi.yaml -g go -o ./gen/client

# Generate TypeScript client
openapi-generator generate -i api/openapi.yaml -g typescript-fetch -o ./gen/ts-client
```

### Validation

Validate the OpenAPI specification:

```bash
# Using openapi-generator
openapi-generator validate -i api/openapi.yaml

# Using spectral (Stoplight)
npx @stoplight/spectral-cli lint api/openapi.yaml
```

## API Overview

### Authentication

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/auth/register` | POST | Register a new user |
| `/auth/login` | POST | Login and get JWT token |

### TODOs

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/todos` | GET | List all TODOs |
| `/todos` | POST | Create a new TODO |
| `/todos/{id}` | GET | Get a specific TODO |
| `/todos/{id}` | PUT | Update a TODO |
| `/todos/{id}` | DELETE | Delete a TODO |

### Health

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Health check |
| `/health/live` | GET | Liveness probe |
| `/health/ready` | GET | Readiness probe |
| `/metrics` | GET | Prometheus metrics |
