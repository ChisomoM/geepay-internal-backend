# Geepay Internal Backend — Quick Start

Get from zero to a running API server in minutes.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Installation](#installation)
3. [Environment Configuration](#environment-configuration)
4. [Running Locally](#running-locally)
5. [Verifying the Server](#verifying-the-server)
6. [Running Tests](#running-tests)
7. [Adding a New Module](#adding-a-new-module)
8. [Adding an Endpoint to an Existing Module](#adding-an-endpoint-to-an-existing-module)
9. [Common Development Workflows](#common-development-workflows)
10. [Troubleshooting](#troubleshooting)
11. [Useful Commands](#useful-commands)

---

## Prerequisites

- **Go** 1.25 or later — `go version` to check
- **Docker** with Docker Compose — `docker --version` to check
- **Make** — available by default on macOS/Linux; on Windows install via Git Bash, WSL, or Chocolatey
- **Git**

---

## Installation

```bash
cd geepay-internal-backend
make setup
```

This runs `go mod download && go mod tidy` to fetch all dependencies.

---

## Environment Configuration

```bash
cp .env.example .env
```

Minimum changes needed for local development:

```env
APP_PORT=8080
APP_ENV=development

# Multi-tenancy — leave enabled for local dev
MULTI_COMPANY_ENABLED=true
DEFAULT_COMPANY_ID=default

# Database — matches docker-compose defaults, no changes needed
DB_HOST=localhost
DB_PORT=5432
DB_NAME=template_db
DB_USERNAME=postgres
DB_PASSWORD=postgres
DB_POOL_MAX_OPEN=25
DB_POOL_MAX_IDLE=5

# Frontend origin for CORS
CORS_ALLOWED_ORIGINS=http://localhost:6050

# Change this in any non-local environment
JWT_SECRET=my-local-dev-secret

# Seeded on first run
SUPER_ADMIN_EMAIL=admin@geepay.com
SUPER_ADMIN_PASSWORD=admin123
```

Email (`SMTP_*`) and file storage (`MINIO_*` or `FILE_STORAGE_TYPE`) are optional for local development. The server starts fine without them.

---

## Running Locally

```bash
make run
```

This single command:
1. Starts PostgreSQL via Docker Compose (`make db-up`)
2. Runs the Go server (`go run main.go`)
3. Auto-runs all migrations on first boot (creates tables)
4. Seeds default data (company record + super admin user)

> **First run:** migrations and seeding take ~10 seconds. Subsequent starts are under 2 seconds.

The API is available at **http://localhost:8080**.

To run the database and server separately:

```bash
make db-up    # just PostgreSQL
go run main.go
```

---

## Verifying the Server

### Health check

```bash
curl http://localhost:8080/api/v1/health
```

Expected: `{"message":"ok"}`

### Login as super admin

```bash
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@geepay.com","password":"admin123"}' | jq .
```

Copy the `access_token` from the response for authenticated calls:

```bash
TOKEN="your_access_token_here"

curl http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer $TOKEN"
```

---

## Running Tests

```bash
make test
```

This runs `go test -v ./...` across all packages.

---

## Adding a New Module

Use the scaffold command first, then fill in the generated files.

### Step 1 — Scaffold

```bash
make new-module NAME=contracts
```

This copies `modules/_template/` into `modules/contracts/` with all `MODULENAME` placeholders replaced.

### Step 2 — Add a database migration

Create `db/migrations/012_contracts_schema.sql`:

```sql
CREATE TABLE IF NOT EXISTS contracts (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id  VARCHAR      NOT NULL,
    name        VARCHAR      NOT NULL,
    status      VARCHAR      NOT NULL DEFAULT 'draft',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_contracts_company_id ON contracts(company_id);
```

### Step 3 — Add the GORM model

Create `models/contract.go` (or add to a relevant existing model file):

```go
package models

type Contract struct {
    CompanyBaseModel
    Name   string `gorm:"not null"`
    Status string `gorm:"default:draft"`
}
```

### Step 4 — Define DTOs in `modules/contracts/models.go`

```go
package contracts

import "time"

type CreateContractRequest struct {
    Name   string `json:"name" validate:"required"`
    Status string `json:"status"`
}

type UpdateContractRequest struct {
    Name   string `json:"name"`
    Status string `json:"status"`
}

type ContractResponse struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    Status    string    `json:"status"`
    CreatedAt time.Time `json:"created_at"`
}
```

### Step 5 — Implement the service in `modules/contracts/service.go`

```go
package contracts

import (
    "geepay-internal-backend/global"
    "geepay-internal-backend/models"
    "gorm.io/gorm"
)

type Service interface {
    List(db *gorm.DB) ([]models.Contract, error)
    Get(db *gorm.DB, id string) (*models.Contract, error)
    Create(db *gorm.DB, req CreateContractRequest) (*models.Contract, error)
    Update(db *gorm.DB, id string, req UpdateContractRequest) (*models.Contract, error)
    Delete(db *gorm.DB, id string) error
}

type service struct {
    app *global.App
}

func NewService(app *global.App) Service {
    return &service{app: app}
}

func (s *service) List(db *gorm.DB) ([]models.Contract, error) {
    var contracts []models.Contract
    result := db.Find(&contracts) // company_id WHERE applied by CompanyMiddleware
    return contracts, result.Error
}

func (s *service) Get(db *gorm.DB, id string) (*models.Contract, error) {
    var contract models.Contract
    result := db.First(&contract, "id = ?", id)
    return &contract, result.Error
}

func (s *service) Create(db *gorm.DB, req CreateContractRequest) (*models.Contract, error) {
    contract := &models.Contract{Name: req.Name, Status: req.Status}
    result := db.Create(contract)
    return contract, result.Error
}

func (s *service) Update(db *gorm.DB, id string, req UpdateContractRequest) (*models.Contract, error) {
    var contract models.Contract
    if err := db.First(&contract, "id = ?", id).Error; err != nil {
        return nil, err
    }
    if req.Name != "" {
        contract.Name = req.Name
    }
    if req.Status != "" {
        contract.Status = req.Status
    }
    result := db.Save(&contract)
    return &contract, result.Error
}

func (s *service) Delete(db *gorm.DB, id string) error {
    return db.Delete(&models.Contract{}, "id = ?", id).Error
}
```

### Step 6 — Implement the handler in `modules/contracts/handler.go`

```go
package contracts

import (
    "net/http"
    "geepay-internal-backend/pkg/response"
    "github.com/labstack/echo/v4"
    "gorm.io/gorm"
)

type Handler struct {
    service Service
}

func NewHandler(svc Service) *Handler {
    return &Handler{service: svc}
}

func (h *Handler) List(c echo.Context) error {
    db := c.Get("db").(*gorm.DB)
    contracts, err := h.service.List(db)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, response.Error("Failed to fetch contracts"))
    }
    return c.JSON(http.StatusOK, response.Success(contracts))
}

func (h *Handler) Get(c echo.Context) error {
    db := c.Get("db").(*gorm.DB)
    id := c.Param("id")
    contract, err := h.service.Get(db, id)
    if err != nil {
        return c.JSON(http.StatusNotFound, response.Error("Contract not found"))
    }
    return c.JSON(http.StatusOK, response.Success(contract))
}

func (h *Handler) Create(c echo.Context) error {
    var req CreateContractRequest
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, response.Error("Invalid request body"))
    }
    db := c.Get("db").(*gorm.DB)
    contract, err := h.service.Create(db, req)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, response.Error("Failed to create contract"))
    }
    return c.JSON(http.StatusCreated, response.Success(contract))
}

func (h *Handler) Update(c echo.Context) error {
    var req UpdateContractRequest
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, response.Error("Invalid request body"))
    }
    db := c.Get("db").(*gorm.DB)
    id := c.Param("id")
    contract, err := h.service.Update(db, id, req)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, response.Error("Failed to update contract"))
    }
    return c.JSON(http.StatusOK, response.Success(contract))
}

func (h *Handler) Delete(c echo.Context) error {
    db := c.Get("db").(*gorm.DB)
    id := c.Param("id")
    if err := h.service.Delete(db, id); err != nil {
        return c.JSON(http.StatusInternalServerError, response.Error("Failed to delete contract"))
    }
    return c.JSON(http.StatusOK, response.SuccessWithMessage("Contract deleted", nil))
}
```

### Step 7 — Register routes in `routes/routes.go`

```go
func SetupContractRoutes(g *echo.Group, h *contracts.Handler, db *gorm.DB, logger *zap.SugaredLogger) {
    g.GET("/contracts",     h.List)
    g.GET("/contracts/:id", h.Get)
    g.POST("/contracts",     h.Create,
        middleware.RBACMiddleware(db, logger, "contracts.create"))
    g.PUT("/contracts/:id",  h.Update,
        middleware.RBACMiddleware(db, logger, "contracts.update"))
    g.DELETE("/contracts/:id", h.Delete,
        middleware.RBACMiddleware(db, logger, "contracts.delete"))
}
```

### Step 8 — Wire into `main.go`

```go
// Initialize service and handler
contractSvc := contracts.NewService(app)
contractHandler := contracts.NewHandler(contractSvc)

// Register routes (inside the protected API group)
routes.SetupContractRoutes(api, contractHandler, db, logger)
```

### Step 9 — Add AutoMigrate entry

In `main.go`, find the `db.AutoMigrate(...)` call and add `&models.Contract{}` to the list.

### Step 10 — Add permissions to seeds (optional)

In `db/seeds.go`, add permission entries:

```go
{Code: "contracts.create", Category: "Contracts", Description: "Create contracts"},
{Code: "contracts.view",   Category: "Contracts", Description: "View contracts"},
{Code: "contracts.update", Category: "Contracts", Description: "Update contracts"},
{Code: "contracts.delete", Category: "Contracts", Description: "Delete contracts"},
```

Restart the server to re-run seeding.

---

## Adding an Endpoint to an Existing Module

1. **Add the DTO** to `modules/<name>/models.go`
2. **Add the service method** to `modules/<name>/service.go`
3. **Add the handler method** to `modules/<name>/handler.go`
4. **Register the route** in `routes/routes.go` under the existing setup function for that module

No changes to `main.go` are needed for additions to existing modules.

---

## Common Development Workflows

### Reading context values in a handler

```go
// Company-scoped database session (set by CompanyMiddleware)
db := c.Get("db").(*gorm.DB)

// Authenticated user info (set by JWTMiddleware)
userID    := c.Get("user_id").(string)
companyID := c.Get("company_id").(string)
roleSlug  := c.Get("role_slug").(string)
```

### Structured logging in a service

```go
s.app.Logger.Infow("Contract created",
    "contract_id", contract.ID,
    "company_id", contract.CompanyID,
)

s.app.Logger.Errorw("Failed to create contract",
    "error", err,
    "request", req,
)
```

Use `Infow`/`Errorw`/`Warnw` with key-value pairs — never `fmt.Printf` or `log.Println`.

### Sending email (optional service)

```go
if s.app.Email != nil {
    if err := s.app.Email.Send(user.Email, "Subject", "Body"); err != nil {
        s.app.Logger.Errorw("Failed to send email", "error", err)
        // Email failure should not fail the main operation
    }
}
```

Always nil-check before calling. Email failures should be logged but not surface as HTTP errors unless the operation is explicitly email-dependent.

### Preloading associations with company scope

Use `Session{NewDB: false}` to preserve the company WHERE clause through Preload calls:

```go
var merchant models.Merchant
db.Session(&gorm.Session{NewDB: false}).
    Preload("Statements").
    First(&merchant, "id = ?", id)
```

Without `NewDB: false`, Preload creates a fresh DB session that loses the company scoping.

### Connecting to the database directly

```
Host:     localhost
Port:     5432
Database: template_db
User:     postgres
Password: postgres
```

Use any PostgreSQL client (psql, TablePlus, DBeaver, etc.).

```bash
psql -h localhost -U postgres -d template_db
```

---

## Troubleshooting

**`dial tcp [::1]:5432: connect: connection refused` at startup**
- PostgreSQL isn't running. Run `make db-up` first, wait a few seconds, then restart the server.

**Migrations fail with `column already exists` or `relation already exists`**
- Never edit an existing migration file. Add a new numbered migration with `IF NOT EXISTS` guards.

**`401 Unauthorized` on all protected endpoints after login**
- `JWT_SECRET` in `.env` must match the secret used when the token was signed. If you changed it, existing tokens are invalid — log in again.
- Access tokens expire in 15 minutes. Use the refresh endpoint or log in again.

**`403 Forbidden` on a specific endpoint**
- The user's role doesn't have the required permission.
- Confirm the permission code in `RBACMiddleware(...)` matches what's in the database.
- Test with the super admin account (`admin@geepay.com`) — super admins bypass all RBAC.

**CORS errors from the frontend**
- `CORS_ALLOWED_ORIGINS` in `.env` must exactly match the frontend's origin, including scheme and port: `http://localhost:6050`.
- Multiple origins are comma-separated: `http://localhost:6050,http://localhost:3000`.

**`make new-module` fails on Windows**
- The Makefile uses `sed` and `cp`. Run it inside Git Bash or WSL. Alternatively, copy `modules/_template/` manually and rename `MODULENAME` occurrences.

**Server starts but returns unexpected data across companies**
- A service method is probably storing `*gorm.DB` on the struct or using a top-level DB reference instead of the one passed per method. Every service method must receive `db *gorm.DB` as a parameter from the handler.

**AutoMigrate doesn't create a new table**
- Make sure the model is added to the `db.AutoMigrate(...)` call in `main.go`.

---

## Useful Commands

| Command | Action |
|---|---|
| `make setup` | Download and tidy Go dependencies |
| `make db-up` | Start PostgreSQL container |
| `make db-down` | Stop PostgreSQL container |
| `make run` | Start database + server |
| `make build` | Compile binary to `bin/server` |
| `make test` | Run all tests with verbose output |
| `make lint` | Run `go fmt` + `go vet` |
| `make new-module NAME=<name>` | Scaffold a new module from template |
| `make clean` | Remove build artifacts |
| `make help` | List all available Makefile targets |
