# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Code Architecture Standards

### File Size Constraints
- **Go files**: Maximum 250 lines per file
- **Directory structure**: Maximum 8 files per folder level (use subfolders if exceeded)

### Code Quality Standards
Avoid these anti-patterns:
1. **Rigidity** - Code difficult to change without cascading modifications
2. **Redundancy** - Repeated code logic across multiple locations  
3. **Circular Dependencies** - Intertwined modules creating maintenance complexity
4. **Fragility** - Changes breaking seemingly unrelated parts
5. **Obscurity** - Unclear intent and chaotic structure
6. **Data Clumps** - Related data items that should be grouped into objects
7. **Unnecessary Complexity** - Over-engineering simple solutions

**Critical**: Always identify and suggest optimization for code smells when encountered.

## Common Commands

### Development Workflow
- `make setup` - Initialize Swagger docs and install dependencies
- `make build` - Build the Go application
- `make test` - Run all tests with race detection and coverage
- `go test ./pkg/{package} -v` - Run tests for specific package
- `make run-env` - Start with .env file configuration (recommended)
- `make run-dev` - Start in development mode
- `make run-local` - Run locally with individual Docker containers
- `make run-local-no-mongo` - Run locally without MongoDB logging

### Environment-Specific Startup
- `make run-dev` - Development environment with debug mode
- `make run-prod` - Production environment with release mode  
- `make run-test` - Test environment configuration

### Docker Operations
- `make up` - Start all services with Docker Compose
- `make up-no-mongo` - Start services without MongoDB
- `make down` - Stop all Docker Compose services
- `make build-docker` - Build Docker images without cache
- `make restart` - Restart Docker Compose services
- `make clean` - Clean up containers and images

### Performance & Testing
- `make test-performance` - Run login API performance tests
- `make test-login` - Simple login endpoint test
- `make test-api-flow` - Complete API test workflow (login → profile → books)
- `make test-all-performance` - Complete performance test suite
- `make debug-jwt TOKEN=<token>` - Debug JWT tokens
- `make db-optimize DB_URL=<url>` - Add database performance indexes

## Architecture Overview

This Go REST API uses Clean Architecture with clear separation of concerns across layers.

### Layered Architecture Flow
```
HTTP Request → Handler → Service → Repository → Database
                ↓         ↓         ↓
               DTO ←→ Models ←→ GORM Models
```

### Core Components

**Application Entry**: `cmd/server/main.go`
- Centralized configuration loading via `pkg/config/config.go`
- Service initialization and dependency injection
- Swagger documentation setup

**Handler Layer**: `pkg/handlers/`
- HTTP request/response handling
- Input validation using `pkg/validators/`
- DTO transformation via `pkg/dto/`
- Standardized responses through `pkg/response/`

**Service Layer**: `pkg/services/`
- Business logic implementation
- Cross-cutting concerns (caching, error handling)
- Transaction management

**Repository Layer**: `pkg/repositories/`
- Data access abstraction
- Database operations using GORM
- Query optimization and indexing

**Infrastructure**: 
- `pkg/database/` - PostgreSQL with GORM, MongoDB for logging
- `pkg/cache/` - Redis caching with interfaces for testability
- `pkg/middleware/` - Authentication, CORS, rate limiting, logging, security
- `pkg/auth/` - JWT token management

### Configuration System
Uses hierarchical configuration loading:
1. `.env` (base configuration)
2. `.env.{environment}` (environment-specific)
3. `.env.local` (local overrides)
4. Environment variables (highest priority)

Key configuration areas:
- **Database**: Connection pooling, timeouts, retry logic
- **Redis**: Connection pooling, timeout configuration  
- **Rate Limiting**: Configurable requests/window
- **Logging**: MongoDB optional, structured logging with Zap
- **Authentication**: JWT secrets, API keys

### API Structure
- **Base Path**: `/api/v1`
- **Authentication**: API key (`X-API-Key`) + JWT tokens (`Authorization: Bearer`)
- **Rate Limiting**: Configurable (default: 60 req/min)
- **Documentation**: Swagger UI at `/swagger/index.html`

### Middleware Chain (Order Matters)
1. Error handling
2. Request context logger  
3. MongoDB status injection
4. Request timeout
5. Performance monitoring
6. Request size limiting
7. Structured logging
8. Security headers (production only)
9. XSS protection (production only)
10. CORS
11. Rate limiting

### Testing Strategy
- **Unit Tests**: Each package with `_test.go` files
- **Mocking**: Interfaces with `golang/mock` generated mocks
- **Performance**: Dedicated scripts in `scripts/` directory
- **E2E**: Python-based tests with pytest
- **Coverage**: Built into `make test` command

### Environment Variables
**Required**:
- `POSTGRES_*` - Database connection
- `JWT_SECRET_KEY`, `API_SECRET_KEY` - Authentication
- `REDIS_HOST` - Caching

**Optional**:
- `MONGO_ENABLED` - Enable/disable MongoDB logging (default: true)
- `RATE_LIMIT_*` - Rate limiting configuration
- `REQUEST_TIMEOUT_MS` - Global request timeout

### Development Patterns

**Error Handling**: Centralized error handling middleware with custom error types in `pkg/apperrors/`

**Validation**: Input validation at handler layer using dedicated validators

**Caching**: Redis-based caching in service layer with cache invalidation strategies

**Logging**: Structured logging with request tracing and optional MongoDB persistence

**Security**: Multiple layers including rate limiting, request size limits, XSS protection, and security headers