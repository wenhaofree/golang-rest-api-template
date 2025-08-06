# CLAUDE.md


This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.
## Code Architecture

- Hard criteria for writing code, including the following principles:

(1) For dynamic languages such as Python, JavaScript, and TypeScript, strive to ensure that each code file does not exceed 200 lines.

(2) For static languages such as Java, Go, and Rust, strive to ensure that each code file does not exceed 250 lines.

(3) For files in each folder level, strive to not exceed 8. If there are more, they need to be planned as multi-level subfolders.

- In addition to the hard criteria, it is also necessary to always pay attention to elegant architecture design and avoid the following "bad smells" that may erode the quality of our code:

(1) Rigidity: The system is difficult to change, and any minor modification can trigger a series of cascading modifications.

(2) Redundancy: The same code logic appears repeatedly in multiple places, making maintenance difficult and prone to inconsistencies.

(3) Circular Dependency: Two or more modules are intertwined, forming an inextricable "knot," making it difficult to test and reuse.

(4) Fragility: A modification to one part of the code leads to unexpected damage to other seemingly unrelated parts of the system.

(5) Obscurity: The code intent is unclear, the structure is chaotic, making it difficult for readers to understand its functionality and design.

(6) Data Clump: Multiple data items always appear together in the parameters of different methods, suggesting that they should be combined into an independent object.

(7) Unnecessary Complexity: Using a "saber" to solve a "chicken" problem, overdesign makes the system bulky and difficult to understand.

- 【Very Important!!】Whether you are writing code yourself, reading, or reviewing others' code, strictly adhere to the above hard criteria and always pay attention to elegant architecture design.

- 【Very Important!!】At any time, once you identify those "bad smells" that may erode the quality of our code, you should immediately ask the user whether optimization is needed and provide reasonable optimization suggestions.

## Common Commands

### Development
- `make setup` - Initialize Swagger documentation and install dependencies
- `make build` - Build the Go application
- `make test` - Run all tests with race detection and coverage
- `make run-local` - Run the application locally with Docker dependencies
- `make up` - Start all services with Docker Compose
- `make down` - Stop all Docker Compose services

### Docker Operations
- `make build-docker` - Build Docker images without cache
- `make restart` - Restart Docker Compose services
- `make clean` - Clean up Docker containers and images

### Testing
- `go test ./... -v` - Run verbose tests
- `go test ./pkg/api -v` - Run tests for specific package
- E2E tests: `cd tests && python e2e.py` (requires Python environment setup)

## Architecture Overview

This is a Go REST API template using the Gin web framework with a layered architecture:

### Core Structure
- **cmd/server/main.go** - Application entry point with Swagger annotations
- **pkg/api/** - HTTP handlers and routing logic
  - `router.go` - Main router setup with middleware chain
  - `books.go` - Book CRUD operations
  - `user.go` - User authentication handlers
- **pkg/models/** - Data models (Book, User)
- **pkg/database/** - Database abstraction layer using GORM
- **pkg/cache/** - Redis caching layer
- **pkg/auth/** - JWT authentication utilities
- **pkg/middleware/** - HTTP middleware (CORS, rate limiting, security, XSS protection)

### Key Technologies
- **Web Framework**: Gin Gonic
- **ORM**: GORM with PostgreSQL
- **Cache**: Redis
- **Logging**: MongoDB + Zap logger
- **Authentication**: JWT tokens
- **Documentation**: Swagger/OpenAPI
- **Testing**: Go testing with mocks

### Database Layer
The application uses a repository pattern with interfaces for testability:
- `Database` interface in `pkg/database/db.go` wraps GORM operations
- `GormDatabase` struct implements the interface
- Models are defined in `pkg/models/` with GORM tags

### API Structure
- Base path: `/api/v1`
- All endpoints require API key authentication (`X-API-Key` header)
- Protected endpoints also require JWT token (`Authorization: Bearer <token>`)
- Rate limiting: 60 requests per minute
- Swagger UI available at `/swagger/index.html`

### Environment Configuration
Required environment variables:
- `POSTGRES_HOST`, `POSTGRES_DB`, `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_PORT`
- `JWT_SECRET_KEY`, `API_SECRET_KEY`
- `REDIS_HOST`

### Testing Strategy
- Unit tests for each package with `_test.go` files
- Mock interfaces generated with `golang/mock`
- E2E tests in Python using pytest
- Test commands include race detection and coverage reporting