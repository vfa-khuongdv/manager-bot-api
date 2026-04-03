# AGENTS.md - Developer Guide for manager-bot/backend

## Project Overview

This is a Go (Golang) CMS project using Gin framework, GORM for MySQL, and follows clean architecture patterns.

## Build & Run Commands

### Development
```bash
# Run with live-reload (recommended)
air

# Run directly
go run cmd/server/main.go

# Run seeder
go run cmd/seeder/seeder.go
```

### Testing
```bash
# Run all tests
go test ./...

# Run single test file
go test -v ./internal/utils/security_test.go

# Run specific test function
go test -v -run TestEncryptPassword ./internal/utils/
```

### Database Migrations
```bash
# Create new migration
migrate create -ext sql -dir internal/database/migrations -seq migration_name

# Apply migrations
migrate -path ./internal/database/migrations -database "mysql://root:root@tcp(127.0.0.1:3306)/golang_db" up
```

### Docker
```bash
docker-compose up --build
```

## Code Style Guidelines

### Architecture Pattern
Follow clean architecture layers:
- **Handlers** (`internal/handlers/`): HTTP request handling, validation
- **Services** (`internal/services/`): Business logic, interface-based
- **Repositories** (`internal/repositories/`): Database operations
- **Models** (`internal/models/`): Data structures with GORM tags

### Interfaces
- Name service interfaces with `I` prefix (e.g., `IProjectService`)
- Define interfaces in same file as implementation
- Constructor functions return interface types

```go
type IProjectService interface {
    GetAll() (*[]models.Project, error)
    GetByID(id uint) (*models.Project, error)
}

type ProjectService struct {
    repo repositories.IProjectRepository
}

func NewProjectService(repo repositories.IProjectRepository) *ProjectService {
    return &ProjectService{repo: repo}
}
```

### Imports
- Standard library first, then external packages, then internal packages
- Group imports with blank line between groups
- Use canonical import paths (e.g., `github.com/vfa-khuongdv/golang-cms/...`)

```go
import (
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    "github.com/vfa-khuongdv/golang-cms/internal/models"
    "github.com/vfa-khuongdv/golang-cms/internal/services"
    "github.com/vfa-khuongdv/golang-cms/internal/utils"
    "github.com/vfa-khuongdv/golang-cms/pkg/errors"
)
```

### Naming Conventions
- **Files**: lowercase with underscores (e.g., `project_handler.go`)
- **Structs**: PascalCase (e.g., `ProjectHandler`)
- **Interfaces**: PascalCase with `I` prefix (e.g., `IProjectService`)
- **Functions/Variables**: camelCase (e.g., `projectService`)
- **Constants**: PascalCase or SCREAMING_SNAKE_CASE for exported (e.g., `ErrServerInternal`)

### Error Handling
Use custom `AppError` from `pkg/errors`:
```go
// Create error
errors.New(errors.ErrDatabaseQuery, "failed to fetch projects")

// With underlying error
errors.Wrap(errors.ErrServerInternal, "operation failed", err)
```

### Response Format
Use utility functions from `internal/utils`:
```go
// Error response
utils.RespondWithError(c, http.StatusBadRequest, errors.New(errors.ErrInvalidData, err.Error()))

// Success response
utils.RespondWithOK(c, http.StatusOK, project)
```

### Model Definition
Use GORM tags for database mapping:
```go
type Project struct {
    ID          int            `json:"id"`
    Name        string         `gorm:"type:varchar(255);not null" json:"name"`
    Description string         `gorm:"type:text" json:"description"`
    Status      string         `gorm:"type:varchar(20);default:'active'" json:"status"`
    SecretKey   string         `gorm:"type:varchar(255)" json:"-"`
    CreatedAt   time.Time      `json:"createdAt"`
    UpdatedAt   time.Time      `json:"updatedAt"`
    DeletedAt   gorm.DeletedAt `json:"deletedAt,omitempty"`
}
```

### Handler Pattern
- Define interface for each handler
- Use dependency injection for services
- Return early on validation errors

```go
func (h *ProjectHandler) GetByID(c *gin.Context) {
    projectId := c.Param("id")
    id, err := strconv.Atoi(projectId)
    if err != nil {
        utils.RespondWithError(c, http.StatusBadRequest, errors.New(errors.ErrInvalidParse, err.Error()))
        return
    }
    // ... rest of handler
}
```

### API Key Authentication
All `/api/v1/*` endpoints require `X-API-Key` header or `api_key` query parameter.

### Environment Configuration
- Load from `.env` file using `godotenv`
- Use `internal/configs/env.go` for configuration loading
- Access via `os.Getenv("VARIABLE_NAME")`

## Project Structure

```
cmd/
  server/main.go       # Main entry point
  seeder/seeder.go     # Database seeder

internal/
  configs/             # Configuration (DB, JWT, env)
  constants/           # Error codes, keys
  database/            # Migrations
  handlers/            # HTTP handlers (v2/ for new APIs)
  middlewares/         # Auth, CORS, logging
  models/              # Data models
  repositories/        # DB access layer
  routes/              # Route definitions
  services/            # Business logic
  utils/               # Helpers (response, validation, security)

pkg/
  errors/              # Custom error types
  logger/              # Logging utility
  mailer/              # Email sending
```

## Testing
- Test files in same package with `_test.go` suffix
- Use `tests/` directory for integration tests
- Follow Go conventions: `func TestXxx(t *testing.T)`

## Common Patterns

### Pagination
Use `utils.Paging` struct for paginated responses.

### Sensitive Data
Use `json:"-"` tag to exclude from JSON serialization:
```go
SecretKey string `gorm:"type:varchar(255)" json:"-"`
```

### Soft Deletes
Use `gorm.DeletedAt` in models - GORM handles automatically.
