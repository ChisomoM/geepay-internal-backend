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
- **PostgreSQL** or **MySQL** running locally — database must be accessible before starting the server
- **Git**

---

## Installation

```bash
cd geepay-internal-backend
go mod download
go mod tidy
```

This fetches all Go dependencies.

---

## Environment Configuration

```bash
cp .env.example .env
```

Edit `.env` with your database connection details and other settings:

```env
APP_PORT=8080
APP_ENV=development

# Multi-tenancy — leave enabled for local dev
MULTI_COMPANY_ENABLED=true
DEFAULT_COMPANY_ID=default

# Database — configure for your local setup (PostgreSQL or MySQL)
# PostgreSQL example:
# DB_HOST=localhost
# DB_PORT=5432
# DB_NAME=geepay_db
# DB_USERNAME=postgres
# DB_PASSWORD=your_password

# MySQL example:
# DB_HOST=localhost
# DB_PORT=3306
# DB_NAME=geepay_db
# DB_USERNAME=root
# DB_PASSWORD=your_password

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

### 1. Start your database

Make sure PostgreSQL or MySQL is running and accessible with the credentials in your `.env` file.

**PostgreSQL** (if installed locally):
```bash
psql -U postgres -d geepay_db
```

**MySQL** (if using MySQL):
```bash
mysql -u root -p
```

### 2. Start the Go server

```bash
go run main.go
```

On first run, the server will:
1. Auto-run all migrations (creates/updates tables)
2. Seed default data (company record + super admin user)

> **First run:** migrations and seeding take ~10 seconds. Subsequent starts are under 2 seconds.

The API is available at **http://localhost:8080**.

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
go test -v ./...
```

This runs tests across all packages.

---

## Adding a New Module

### Step 1 — Scaffold

Copy `modules/_template/` to `modules/contracts/` (or your module name):

```bash
cp -r modules/_template modules/contracts
```

Then manually replace all occurrences of `MODULENAME` in the files with your module name (e.g., `contracts`).

### Step 2 — Add the GORM model

Create `models/contract.go` (or add to a relevant existing model file):

```go
package models

type Contract struct {
    CompanyBaseModel
    Name   string `gorm:"not null"`
    Status string `gorm:"default:draft"`
}
```

### Step 3 — Define DTOs in `modules/contracts/models.go`

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

### Step 4 — Implement the service in `modules/contracts/service.go`

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

### Step 5 — Implement the handler in `modules/contracts/handler.go`

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

### Step 6 — Register routes in `routes/routes.go`

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

### Step 7 — Wire into `main.go`

```go
// Initialize service and handler
contractSvc := contracts.NewService(app)
contractHandler := contracts.NewHandler(contractSvc)

// Register routes (inside the protected API group)
routes.SetupContractRoutes(api, contractHandler, db, logger)
```

### Step 8 — Add AutoMigrate entry

In `main.go`, find the `db.AutoMigrate(...)` call and add `&models.Contract{}` to the list. This automatically creates/updates the table schema on startup.

### Step 9 — Add permissions to seeds (optional)

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

## Troubleshooting

**`dial tcp localhost:5432: connect: connection refused` or MySQL connection error at startup**
- The database isn't running. Start PostgreSQL or MySQL and ensure it's accessible with your `.env` credentials. Then restart the server.

**AutoMigrate creates or updates tables**
- AutoMigrate uses GORM models to automatically create/update table schemas on startup.
- If you modify a model (add/remove/change a field), the schema updates automatically on next server start.
- Use GORM tags like `gorm:"column:name"` to control field mapping.

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

**Module scaffolding on Windows**
- Copy `modules/_template/` manually to your new module folder, then find-and-replace all `MODULENAME` occurrences with your actual module name.

**Server starts but returns unexpected data across companies**
- A service method is probably storing `*gorm.DB` on the struct or using a top-level DB reference instead of the one passed per method. Every service method must receive `db *gorm.DB` as a parameter from the handler.

**AutoMigrate doesn't create a new table**
- Make sure the model is added to the `db.AutoMigrate(...)` call in `main.go`.

---

## Useful Commands

| Command | Action |
|---|---|
| `go mod download && go mod tidy` | Download and tidy Go dependencies |
| `go run main.go` | Start the server |
| `go build -o bin/server` | Compile binary to `bin/server` |
| `go test -v ./...` | Run all tests with verbose output |
| `go fmt ./... && go vet ./...` | Format and vet code |
| `cp -r modules/_template modules/<name>` | Copy template module (then manually replace MODULENAME) |
| `rm -rf bin/` | Remove build artifacts |
