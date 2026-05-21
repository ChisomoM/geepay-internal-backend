# Template Usage Guide

This guide walks through using this template for a new project, explains the three architectural patterns, and clarifies what to keep vs. what to replace.

---

## Phase 0: Clone and Initialize (15 minutes)

### 1. Clone the template

```bash
git clone <template-repo> my-project
cd my-project
rm -rf .git  # Detach from template repo
git init     # Start fresh repo
```

### 2. Update module declaration

Replace `backend` with your project name in `go.mod`:

```go
// go.mod
module github.com/yourorg/my-project/backend

go 1.21
```

### 3. Configure environment

```bash
cp .env.example .env

# Edit .env with your database credentials
# Minimum required:
#   DB_HOST
#   DB_PORT
#   DB_NAME
#   JWT_SECRET
```

### 4. Verify setup

```bash
make db-up
go run main.go
```

Visit `http://localhost:8080/api/v1/health` — you should see `{"status": "ok"}`.

---

## Phase 1: Understand the Three Architectural Patterns

This template uses three patterns for modules. Understanding the differences is critical to maintaining consistency.

### Pattern 1: New Module (Recommended for domain features)

**Use when:** Building a new feature that doesn't exist in the template.

**Location:** `modules/yourdomain/` (e.g., `modules/projects/`, `modules/invoices/`)

**Structure:**

```
modules/projects/
├── service.go     # Business logic: interface + implementation
├── handler.go     # HTTP handlers
├── models.go      # Request/response DTOs
└── public.go      # Exported interfaces
```

**Key characteristics:**
- Service receives `*global.App` in constructor (infrastructure only)
- Service methods receive `*gorm.DB` as first parameter (tenant-scoped by middleware)
- Handler extracts `db` from Echo context: `db := c.Get("db").(*gorm.DB)`
- No circular dependencies (modules import interfaces, not other modules)

**Example:**

```go
// modules/projects/service.go
type Service interface {
    Create(db *gorm.DB, req CreateRequest) (*ProjectDTO, error)
    List(db *gorm.DB) ([]ProjectDTO, error)
}

type service struct {
    app *global.App
}

func (s *service) Create(db *gorm.DB, req CreateRequest) (*ProjectDTO, error) {
    project := &models.Project{
        TenantID: tenantID,  // Automatically scoped
        Name:     req.Name,
    }
    return db.Create(project).Error
}

// modules/projects/handler.go
type Handler interface {
    Create(c echo.Context) error
    List(c echo.Context) error
}

type handler struct {
    service Service
}

func (h *handler) Create(c echo.Context) error {
    db := c.Get("db").(*gorm.DB)
    req := &CreateRequest{}
    c.BindJSON(req)
    
    dto, err := h.service.Create(db, req)
    if err != nil {
        return c.JSON(http.StatusBadRequest, response.Error("Failed"))
    }
    return c.JSON(http.StatusCreated, response.Success(dto))
}
```

**When to use:**
- ✅ New CRUD features
- ✅ New business logic
- ✅ New integrations
- ❌ Don't use for infrastructure (use global.App instead)

---

### Pattern 2: Thin Wrapper Module (For legacy code migration)

**Use when:** You have legacy handler/service pairs that you want to expose through the module pattern while migrating incrementally.

**Location:** `modules/yourdomain/handler.go` (one file, delegates to legacy handler)

**Structure:**

```go
// modules/documents/handler.go
type Handler interface { ... }

type handler struct {
    legacy *handlers.DocumentHandler
}

func NewHandler(h *handlers.DocumentHandler) Handler {
    return &handler{legacy: h}
}

func (h *handler) Create(c echo.Context) error {
    return h.legacy.Create(c)  // Delegate to legacy handler
}
```

**When to use:**
- ✅ Exposing legacy handlers through the module interface
- ✅ Gradual migration from handlers/ to modules/
- ❌ Don't use for new features (use Pattern 1 instead)

---

### Pattern 3: Infrastructure Module (Shared across all modules)

**Use when:** Building services that all modules depend on (email, logging, file upload, auth).

**Location:** `global/app.go` as interfaces + `services/` as implementations

**Structure:**

```go
// global/app.go
type App struct {
    Config *config.Config
    Logger *zap.SugaredLogger
    Email  EmailSender    // Interface, not concrete type
    Upload FileUploader   // Interface, not concrete type
}

// Interfaces defined in global/
type EmailSender interface {
    Send(to, subject, body string) error
    SendTemplate(to, templateName string, data map[string]interface{}) error
}

type FileUploader interface {
    Upload(bucket, name string, data []byte) (url string, err error)
}
```

**When to use:**
- ✅ Infrastructure: logging, config, email, file storage
- ✅ Cross-cutting concerns
- ❌ Don't use for domain logic (use module pattern instead)

---

## Phase 2: Add Your First Feature

### Step-by-step example: Create a "Projects" module

#### Step 1: Create the directory structure

```bash
make new-module NAME=projects
```

Or manually:

```bash
mkdir -p modules/projects
cp -r modules/_template/* modules/projects/
```

#### Step 2: Create the domain model

```go
// models/project.go
package models

type Project struct {
    TenantBaseModel
    Name        string `json:"name" gorm:"index"`
    Description string `json:"description"`
    OwnerID     uuid.UUID `json:"owner_id" gorm:"type:uuid"`
    IsActive    bool `json:"is_active" gorm:"default:true"`
}
```

#### Step 3: Implement the service

```go
// modules/projects/service.go
package projects

import (
    "backend/global"
    "backend/models"
    "gorm.io/gorm"
)

type CreateRequest struct {
    Name        string `json:"name" binding:"required"`
    Description string `json:"description"`
}

type ProjectDTO struct {
    ID          uuid.UUID `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    CreatedAt   time.Time `json:"created_at"`
}

type Service interface {
    Create(db *gorm.DB, userID string, req CreateRequest) (*ProjectDTO, error)
    List(db *gorm.DB) ([]ProjectDTO, error)
    Get(db *gorm.DB, id string) (*ProjectDTO, error)
}

type service struct {
    app *global.App
}

func NewService(app *global.App) Service {
    return &service{app: app}
}

func (s *service) Create(db *gorm.DB, userID string, req CreateRequest) (*ProjectDTO, error) {
    project := &models.Project{
        TenantBaseModel: models.TenantBaseModel{
            TenantID: tenantID,  // Extract from context
        },
        Name:        req.Name,
        Description: req.Description,
        OwnerID:     uuid.MustParse(userID),
    }
    
    if err := db.Create(project).Error; err != nil {
        s.app.Logger.Errorf("Failed to create project: %v", err)
        return nil, err
    }
    
    return &ProjectDTO{
        ID:          project.ID,
        Name:        project.Name,
        Description: project.Description,
        CreatedAt:   project.CreatedAt,
    }, nil
}

func (s *service) List(db *gorm.DB) ([]ProjectDTO, error) {
    var projects []models.Project
    if err := db.Find(&projects).Error; err != nil {
        return nil, err
    }
    
    var dtos []ProjectDTO
    for _, p := range projects {
        dtos = append(dtos, ProjectDTO{
            ID:        p.ID,
            Name:      p.Name,
            CreatedAt: p.CreatedAt,
        })
    }
    return dtos, nil
}

func (s *service) Get(db *gorm.DB, id string) (*ProjectDTO, error) {
    var project models.Project
    if err := db.First(&project, "id = ?", id).Error; err != nil {
        return nil, err
    }
    
    return &ProjectDTO{
        ID:          project.ID,
        Name:        project.Name,
        Description: project.Description,
        CreatedAt:   project.CreatedAt,
    }, nil
}
```

#### Step 4: Implement the handler

```go
// modules/projects/handler.go
package projects

import (
    "backend/pkg/response"
    "net/http"
    "github.com/labstack/echo/v4"
    "gorm.io/gorm"
)

type Handler interface {
    Create(c echo.Context) error
    List(c echo.Context) error
    Get(c echo.Context) error
}

type handler struct {
    service Service
}

func NewHandler(service Service) Handler {
    return &handler{service: service}
}

func (h *handler) Create(c echo.Context) error {
    db := c.Get("db").(*gorm.DB)
    userID := c.Get("user_id").(string)
    
    var req CreateRequest
    if err := c.BindJSON(&req); err != nil {
        return c.JSON(http.StatusBadRequest, response.Error("Invalid request"))
    }
    
    dto, err := h.service.Create(db, userID, req)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, response.Error("Failed to create project"))
    }
    
    return c.JSON(http.StatusCreated, response.SuccessWithMessage("Project created", dto))
}

func (h *handler) List(c echo.Context) error {
    db := c.Get("db").(*gorm.DB)
    
    dtos, err := h.service.List(db)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, response.Error("Failed to list projects"))
    }
    
    return c.JSON(http.StatusOK, response.Success(dtos))
}

func (h *handler) Get(c echo.Context) error {
    db := c.Get("db").(*gorm.DB)
    id := c.Param("id")
    
    dto, err := h.service.Get(db, id)
    if err != nil {
        if err == gorm.ErrRecordNotFound {
            return c.JSON(http.StatusNotFound, response.Error("Project not found"))
        }
        return c.JSON(http.StatusInternalServerError, response.Error("Failed to get project"))
    }
    
    return c.JSON(http.StatusOK, response.Success(dto))
}
```

#### Step 5: Create migration

```sql
-- db/migrations/004_projects.sql
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id VARCHAR(255) NOT NULL REFERENCES tenants(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    owner_id UUID NOT NULL REFERENCES users(id),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_projects_tenant_id ON projects(tenant_id);
CREATE INDEX idx_projects_owner_id ON projects(owner_id);

ALTER TABLE projects ENABLE ROW LEVEL SECURITY;

CREATE POLICY projects_tenant_isolation ON projects
    USING (tenant_id = current_setting('app.tenant_id')::text)
    WITH CHECK (tenant_id = current_setting('app.tenant_id')::text);
```

#### Step 6: Register routes

```go
// main.go
projectsService := projects.NewService(app)
projectsHandler := projects.NewHandler(projectsService)

// routes/routes.go
func SetupProjectRoutes(g *echo.Group, h projects.Handler) {
    g.POST("/projects", h.Create, middleware.RBACMiddleware(db.GetDB(), logger, "projects.create"))
    g.GET("/projects", h.List)
    g.GET("/projects/:id", h.Get)
}

// In main.go, register the route group
routes.SetupProjectRoutes(api, projectsHandler)
```

#### Step 7: Add permissions (seed)

```go
// In main.go or db/migrations/
db.SeedDefaultRoles(migrationDB, tenantID, sugar)

// Then manually insert project permissions:
// INSERT INTO permissions (id, tenant_id, code, category, description) 
// VALUES (gen_random_uuid(), 'your-tenant', 'projects.create', 'projects', 'Create projects'),
//        (gen_random_uuid(), 'your-tenant', 'projects.export', 'projects', 'Export projects');
```

Done! Your Projects module is now fully integrated with multi-tenancy, RBAC, and audit logging.

---

## Phase 3: What NOT to Do

### ❌ Don't couple modules to each other

```go
// BAD: module importing another module's concrete types
import "backend/modules/documents"

func (s *service) CreateWithDocument(db *gorm.DB, docID string) error {
    doc := documents.Get(db, docID)  // Circular dependency risk
}

// GOOD: Use interfaces and dependency injection
type DocumentLoader interface {
    Get(db *gorm.DB, id string) (*DocumentDTO, error)
}

func (s *service) CreateWithDocument(db *gorm.DB, dl DocumentLoader, docID string) error {
    doc := dl.Get(db, docID)  // Injected, no coupling
}
```

### ❌ Don't store DB in service constructor

```go
// BAD: Service holds a reference to the database
type service struct {
    db *gorm.DB  // Wrong! Forces single-tenancy or global state
}

func (s *service) Create(req CreateRequest) error {
    s.db.Create(...)  // Which tenant?
}

// GOOD: Service receives DB as method parameter
type service struct {
    app *global.App  // Infrastructure only
}

func (s *service) Create(db *gorm.DB, req CreateRequest) error {
    db.Create(...)  // DB is already tenant-scoped by middleware
}
```

### ❌ Don't skip RLS policies

```sql
-- BAD: Table without RLS
CREATE TABLE documents (
    id UUID PRIMARY KEY,
    tenant_id VARCHAR(255),
    name VARCHAR(255)
);

-- GOOD: Table with RLS
CREATE TABLE documents (
    id UUID PRIMARY KEY,
    tenant_id VARCHAR(255),
    name VARCHAR(255)
);

ALTER TABLE documents ENABLE ROW LEVEL SECURITY;

CREATE POLICY documents_tenant_isolation ON documents
    USING (tenant_id = current_setting('app.tenant_id')::text)
    WITH CHECK (tenant_id = current_setting('app.tenant_id')::text);
```

### ❌ Don't bypass middleware

```go
// BAD: Creating a custom route without standard middleware
e.GET("/special-endpoint", func(c echo.Context) error {
    // No JWT validation, no tenant scoping!
    return c.JSON(200, "unprotected data")
})

// GOOD: Route uses standard middleware chain
api.GET("/special-endpoint", handler.Get)  // Inherits JWT + TenantMiddleware
```

### ❌ Don't hardcode tenant_id

```go
// BAD: Service assumes a hardcoded tenant
func (s *service) ListAll(db *gorm.DB) ([]ProjectDTO, error) {
    var projects []models.Project
    db.Where("tenant_id = ?", "hardcoded-tenant-id").Find(&projects)
    return ...
}

// GOOD: Middleware handles tenant scoping
func (s *service) ListAll(db *gorm.DB) ([]ProjectDTO, error) {
    var projects []models.Project
    db.Find(&projects)  // DB is already WHERE tenant_id = ? scoped
    return ...
}
```

---

## Phase 4: Configuration Decisions

### Single-Tenant vs Multi-Tenant

**Multi-tenant (default):**

```bash
# .env
MULTI_TENANT_ENABLED=true
DEFAULT_TENANT_ID=your-tenant-uuid
```

All queries scoped by tenant_id. Add new tenants by inserting into the `tenants` table.

**Single-tenant:**

```bash
# .env
MULTI_TENANT_ENABLED=false
DEFAULT_TENANT_ID=default
```

TenantMiddleware becomes a no-op. No WHERE tenant_id clauses needed.

### Email Configuration

**No email (default):**

```go
// main.go
var emailService global.EmailSender // nil
app := global.New(cfg, sugar, emailService, uploadService)

// Usage in modules: guard against nil
if app.Email != nil {
    app.Email.Send(...)
}
```

**SMTP:**

```bash
# .env
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM_EMAIL=noreply@yourapp.com
```

Implement and inject in main.go:

```go
emailService := services.NewSMTPSender(cfg)
app := global.New(cfg, sugar, emailService, uploadService)
```

### File Storage Configuration

**Local filesystem:**

```bash
FILE_STORAGE_TYPE=local
```

```go
uploadService := &services.LocalUploader{BasePath: "/tmp/uploads"}
```

**MinIO (S3-compatible):**

```bash
FILE_STORAGE_TYPE=minio
MINIO_URL=http://localhost:9000
MINIO_BUCKET=myapp
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_PUBLIC_URL=http://localhost:9000
```

```go
uploadService := services.NewMinIOUploader(cfg)
```

---

## Phase 5: Extending Core Infrastructure

### Adding a new infrastructure service

Example: Cache service

#### Step 1: Define interface in `global/app.go`

```go
// global/app.go
type Cacher interface {
    Get(key string) (interface{}, error)
    Set(key string, value interface{}, ttl time.Duration) error
    Delete(key string) error
}
```

#### Step 2: Add to App struct

```go
type App struct {
    Config *config.Config
    Logger *zap.SugaredLogger
    Email  EmailSender
    Upload FileUploader
    Cache  Cacher  // New
}
```

#### Step 3: Implement concrete type

```go
// services/redis_cache.go
type RedisCache struct {
    client *redis.Client
}

func (c *RedisCache) Get(key string) (interface{}, error) {
    return c.client.Get(context.Background(), key).Result()
}

func (c *RedisCache) Set(key string, value interface{}, ttl time.Duration) error {
    return c.client.Set(context.Background(), key, value, ttl).Err()
}
```

#### Step 4: Inject in main.go

```go
cache := &services.RedisCache{client: redis.NewClient(...)}
app := &global.App{
    Config: cfg,
    Logger: sugar,
    Email:  emailService,
    Upload: uploadService,
    Cache:  cache,
}
```

#### Step 5: Use in modules

```go
// modules/projects/service.go
func (s *service) Get(db *gorm.DB, id string) (*ProjectDTO, error) {
    // Try cache first
    if cached, _ := s.app.Cache.Get("project:" + id); cached != nil {
        return cached.(*ProjectDTO), nil
    }
    
    // Query database
    project := &models.Project{}
    db.First(project, id)
    
    // Store in cache
    s.app.Cache.Set("project:" + id, project, 1*time.Hour)
    
    return &ProjectDTO{...}, nil
}
```

---

## Phase 6: Testing

### Unit test example

```go
// modules/projects/service_test.go
package projects

import (
    "testing"
    "backend/models"
    "gorm.io/gorm"
)

func TestCreateProject(t *testing.T) {
    // Setup mock app
    app := &global.App{
        Logger: zap.NewNop().Sugar(),
    }
    
    // Create service
    svc := &service{app: app}
    
    // Create test request
    req := CreateRequest{
        Name:        "Test Project",
        Description: "Test",
    }
    
    // Use in-memory SQLite for testing (or mock DB)
    db := setupTestDB()
    defer db.Close()
    
    // Test
    dto, err := svc.Create(db, "user-123", req)
    if err != nil {
        t.Fatalf("Expected no error, got %v", err)
    }
    
    if dto.Name != "Test Project" {
        t.Errorf("Expected name to be 'Test Project', got %s", dto.Name)
    }
}
```

### Integration test example

```go
// tests/projects_test.go
package tests

import (
    "testing"
    "bytes"
    "encoding/json"
    "net/http/httptest"
)

func TestCreateProjectEndpoint(t *testing.T) {
    // Setup test server
    e := setupTestServer()
    
    // Create request
    body := map[string]string{
        "name": "Test Project",
    }
    bodyBytes, _ := json.Marshal(body)
    
    // Execute request
    req := httptest.NewRequest("POST", "/api/v1/projects", bytes.NewReader(bodyBytes))
    req.Header.Set("Authorization", "Bearer "+testToken)
    req.Header.Set("Content-Type", "application/json")
    
    rec := httptest.NewRecorder()
    e.ServeHTTP(rec, req)
    
    // Assert
    if rec.Code != 201 {
        t.Errorf("Expected status 201, got %d", rec.Code)
    }
}
```

---

## Checklist: Launching Your Project

- [ ] Clone template and initialize new repo
- [ ] Update `go.mod` module name
- [ ] Configure `.env` with database credentials
- [ ] Create domain models in `models/`
- [ ] Create first module using `make new-module`
- [ ] Implement service and handler
- [ ] Create migration file for new schema
- [ ] Register routes in `routes/routes.go`
- [ ] Test endpoints with curl or Postman
- [ ] Add audit logging (automatic via AuditMiddleware)
- [ ] Add RBAC permissions (manual seed or migration)
- [ ] Write unit tests
- [ ] Write integration tests
- [ ] Deploy to staging
- [ ] Review logs for errors
- [ ] Launch to production

---

## FAQ

### Q: Can I use this template for a non-SaaS project?
**A:** Yes. Set `MULTI_TENANT_ENABLED=false` in `.env`. The template still works, just without multi-tenancy.

### Q: What if I don't need all the models (User, Role, Permission)?
**A:** The template provides them because RBAC is useful for almost every project. If you truly don't need them, delete them after cloning. But it's safer to keep them and just not use them.

### Q: Should I version the migrations?
**A:** Yes. SQL migrations in `db/migrations/` should be versioned (e.g., `001_core_schema.sql`, `002_users.sql`). Never modify existing migrations — always create new ones.

### Q: How do I handle multiple environments (dev, staging, prod)?
**A:** Use separate `.env` files for each environment and pass the config file path at startup. Or use environment-specific flags (e.g., `APP_ENV=production`).

### Q: Can I add caching?
**A:** Yes. Define a `Cacher` interface in `global/app.go` and inject a Redis or in-memory implementation. See "Extending Core Infrastructure" above.

### Q: How do I handle file uploads?
**A:** Use the `FileUploader` interface. Implement for local filesystem, MinIO, AWS S3, etc., and inject in `main.go`.

### Q: What if I need to schedule tasks?
**A:** Use a library like `go-co-op/gocron` for background jobs. Inject the scheduler into modules that need it.

### Q: How do I run migrations in production?
**A:** Migrations run automatically on startup. For zero-downtime deployments, consider running migrations manually before deploying the new code.

---

## Support

Questions? Issues? Open an issue on GitHub or reach out to the team.
