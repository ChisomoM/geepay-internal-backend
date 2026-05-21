# Avec Go Backend Template

A production-ready Go backend template using Echo, GORM, and PostgreSQL with built-in multi-tenancy support, RBAC, and modular architecture.

**Design Philosophy:** Clean, maintainable, scalable, with a focus on developer experience and security.

---

## Quick Start

### Prerequisites
- Go 1.21+
- PostgreSQL 14+
- Make (optional, for Makefile commands)

### Setup (5 minutes)

```bash
# 1. Clone template
git clone <template-repo> my-project
cd my-project

# 2. Create .env from example
cp .env.example .env

# 3. Update .env with your database credentials
# Required vars: DB_HOST, DB_PORT, DB_NAME, JWT_SECRET

# 4. Start PostgreSQL
docker-compose up -d postgres

# 5. Run migrations and start server
go run main.go
```

The server starts on `http://localhost:8080` with default endpoints:
- `GET /api/v1/health` → `{"status": "ok"}`
- `POST /api/v1/auth/login` → Authenticate
- `GET /api/v1/users` → List users (requires auth)

---

## Architecture Overview

### Core Layers

```
main.go
  ↓
Middleware (JWT, Tenant, RBAC, Audit)
  ↓
Handler (HTTP layer)
  ↓
Service (Business logic, receives *gorm.DB)
  ↓
Models (Domain entities)
  ↓
Database (Tenant-scoped via middleware)
```

### Key Components

| Component | Purpose | Location |
|---|---|---|
| **global.App** | Infrastructure holder (Config, Logger, Email, Upload) | `global/app.go` |
| **Middleware** | Request processing pipeline (Auth, Tenancy, RBAC, Audit) | `middleware/` |
| **Modules** | Feature-specific handler/service pairs | `modules/` |
| **Models** | Domain entities with GORM tags | `models/` |
| **DB Manager** | Singleton connection pool + migrations | `db/` |

---

## Multi-Tenancy

### How It Works

1. **JWT contains `tenant_id`** — Set during login
2. **TenantMiddleware scopes queries** — `WHERE tenant_id = ?` at request middleware level
3. **No per-query effort** — Service methods receive tenant-scoped DB automatically
4. **Row-Level Security** — PostgreSQL RLS policies provide defense-in-depth

### Single-Tenant Mode

For non-SaaS projects, disable multi-tenancy:

```bash
# .env
MULTI_TENANT_ENABLED=false
DEFAULT_TENANT_ID=default
```

TenantMiddleware becomes a no-op; all queries work on the full database.

---

## Dual Authentication System

This template uses **two separate authentication contexts**:

| Aspect | Tenant Auth | ControlHub Auth |
|---|---|---|
| **Purpose** | User login within a tenant | Platform admin login |
| **Endpoint** | `POST /api/v1/auth/login` | `POST /controlhub/auth/login` |
| **JWT Claims** | `{sub, email, role_slug, tenant_id}` | `{sub, email, role, level, user_type}` |
| **DB Scope** | Tenant-scoped (WHERE tenant_id) | Platform-scoped (all rows) |
| **Table** | `users` | `company_admins` |
| **RBAC** | Checked via role_slug + permissions table | Checked via role/level claims |
| **Use Case** | Application users | Infrastructure admins |

**Why Two Systems?**
- **Separation of Concerns** — Platform management (infrastructure) is separate from application data (users/data)
- **Security** — Breached tenant credentials can't access platform
- **Scalability** — Each context has its own token lifecycle and permissions model

---

## Adding a New Feature (Module)

### Step 1: Create module directory

```bash
mkdir -p modules/projects
```

### Step 2: Copy template scaffold

```bash
cp -r modules/_template/* modules/projects/
```

### Step 3: Update package names and business logic

Replace `MODULENAME` with `projects` in all files.

### Step 4: Implement service methods

```go
// modules/projects/service.go
func (s *service) Create(db *gorm.DB, req CreateRequest) (*ResponseDTO, error) {
    project := &models.Project{
        TenantID: tenantID,  // Automatically scoped by middleware
        Name: req.Name,
        // ... other fields
    }
    return db.Create(project).Error
}
```

### Step 5: Register routes

```go
// main.go
projectsHandler := projects.NewHandler(projects.NewService(app))
routes.SetupProjectRoutes(api, projectsHandler)

// routes/routes.go
func SetupProjectRoutes(g *echo.Group, h projects.Handler) {
    g.POST("/projects", h.Create, middleware.RBACMiddleware(db.GetDB(), logger, "projects.create"))
    g.GET("/projects", h.List)
    // ...
}
```

---

## Configuration

### Required Environment Variables

```bash
# Application
APP_PORT=8080
APP_ENV=development

# Multi-Tenancy
MULTI_TENANT_ENABLED=true
DEFAULT_TENANT_ID=your-tenant-id

# Database (two-role strategy for safety)
DB_HOST=localhost
DB_PORT=5432
DB_NAME=template_db
DB_USERNAME=postgres
DB_PASSWORD=password

# Database two-role credentials (optional, falls back to above)
DB_APP_USERNAME=app_user          # Runtime queries (no BYPASSRLS)
DB_APP_PASSWORD=app_password
DB_MIGRATION_USERNAME=migration_user  # Migrations only (BYPASSRLS)
DB_MIGRATION_PASSWORD=migration_pass

# Connection pool
DB_POOL_MAX_OPEN=25
DB_POOL_MAX_IDLE=5
DB_STATEMENT_TIMEOUT_MS=30000

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173

# JWT
JWT_SECRET=your-secret-key-change-in-production

# Optional: Email (if using SMTP)
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USERNAME=user
SMTP_PASSWORD=pass
SMTP_FROM_EMAIL=noreply@example.com

# Optional: File storage
FILE_STORAGE_TYPE=local  # or minio, s3
MINIO_URL=http://localhost:9000
MINIO_BUCKET=documents
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
```

See `.env.example` for all options.

---

## Database Schema

### Two-Role Strategy

The template uses two PostgreSQL roles for safety:

1. **dims_app_user** (no BYPASSRLS) — All runtime queries
   - Cannot modify schema
   - Enforces RLS policies
   - Safe if compromised

2. **dims_migration_user** (BYPASSRLS) — Schema changes only
   - Used at startup for migrations
   - Closed immediately after
   - Never used at runtime

### Core Tables

```sql
-- Tenants (multi-tenant SaaS organization)
tenants (id, name, slug, is_active, ...)

-- Users (scoped to tenant)
users (id, tenant_id, email, password_hash, role_slug, ...)

-- Roles (scoped to tenant)
roles (id, tenant_id, name, slug, is_system, ...)

-- Permissions (scoped to tenant)
permissions (id, tenant_id, code, category, ...)

-- role_permissions (join table)
role_permissions (role_id, permission_id)

-- User Permission Overrides (fine-grained per-user control)
user_permission_overrides (id, user_id, permission_id, granted, ...)

-- Audit logs (scoped to tenant)
audit_logs (id, tenant_id, user_id, action, resource, ...)
```

### Adding New Tables

1. Create migration file in `db/migrations/`

```sql
-- db/migrations/004_projects.sql
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id VARCHAR(255) NOT NULL REFERENCES tenants(id),
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    -- ... other fields
);

CREATE INDEX idx_projects_tenant_id ON projects(tenant_id);
ALTER TABLE projects ENABLE ROW LEVEL SECURITY;

CREATE POLICY projects_tenant_isolation ON projects
    USING (tenant_id = current_setting('app.tenant_id')::text)
    WITH CHECK (tenant_id = current_setting('app.tenant_id')::text);
```

2. Create model in `models/projects.go`

```go
type Project struct {
    TenantBaseModel
    Name  string
    // ... other fields
}
```

3. Run migrations at startup (already automatic in `main.go`)

---

## RBAC (Role-Based Access Control)

### Permission Model

Permissions are defined as `resource.action` codes:

```
users.view      — View users
users.create    — Create users
users.update    — Update users
users.delete    — Delete users
projects.export — Export projects
admin.manage    — Admin panel access
```

### Default Roles

- `super_admin` — Bypass all permission checks
- `admin` — Full access to admin permissions
- `user` — Limited to basic permissions
- Custom roles — Define per tenant

### Fine-Grained Overrides

For exceptions, use `UserPermissionOverride`:

```go
// Grant a user extra permission
override := UserPermissionOverride{
    UserID: userID,
    PermissionID: permissionID,
    Granted: true,
    Reason: "Special case: temporary export access",
}
db.Create(&override)

// Deny a permission despite role having it
override := UserPermissionOverride{
    UserID: userID,
    PermissionID: permissionID,
    Granted: false,  // Revoke
    Reason: "Contract ended",
}
db.Create(&override)
```

### Protecting Routes

```go
// Require specific permission
api.POST("/export", handler.Export, middleware.RBACMiddleware(db, logger, "projects.export"))

// Require admin-only
api.POST("/users", handler.Create, middleware.RBACMiddleware(db, logger, "admin.manage"))
```

---

## Middleware Stack

Middleware runs in this order:

1. **LoggerMiddleware** — Structured request logging via Zap
2. **CORSMiddleware** — CORS headers (config-driven)
3. **JWTMiddleware** — Token validation, claims extraction
4. **TenantMiddleware** — Tenant resolution, DB scoping
5. **AuditMiddleware** — State-change logging (POST, PUT, DELETE)
6. **RBACMiddleware** — Per-route permission checks (optional)

---

## API Response Format

All endpoints return a consistent format:

```json
{
  "status": 0,
  "message": "success",
  "data": { /* payload */ }
}
```

**Status codes:**
- `0` = Success
- `1` = Error

**Response helpers:**

```go
// Success
response.Success(data)
response.SuccessWithMessage("User created", userDTO)

// Error
response.Error("Invalid email")
response.ErrorWithData("Validation failed", errors)
```

---

## Authentication Flow (Tenant-Level)

### 1. Login

```bash
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}
```

Response:

```json
{
  "status": 0,
  "message": "success",
  "data": {
    "access_token": "eyJhbGc...",
    "refresh_token": "refresh...",
    "user": {
      "id": "uuid",
      "email": "user@example.com",
      "role_slug": "admin"
    }
  }
}
```

### 2. Use access token

```bash
GET /api/v1/users
Authorization: Bearer eyJhbGc...
```

### 3. Refresh token (after expiry)

```bash
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "refresh..."
}
```

---

## ControlHub Authentication (Platform-Level)

ControlHub provides **company admin** authentication separate from tenant-level user auth. This is the platform administration layer.

### 1. ControlHub Login (Company Admin)

```bash
POST /controlhub/auth/login
Content-Type: application/json

{
  "email": "admin@company.com",
  "password": "secure_password"
}
```

Response:

```json
{
  "status": 0,
  "message": "success",
  "data": {
    "access_token": "eyJhbGc...",
    "refresh_token": "refresh...",
    "email": "admin@company.com",
    "role": "controlhub",
    "user_type": "controlhub_admin"
  }
}
```

**Key Differences from Tenant Auth:**
- **No tenant_id in JWT** — Platform-scoped, not tenant-specific
- **Role: "controlhub"** — Indicates platform admin, not tenant user
- **Access Level: "company_admin"** — Specifies admin privileges
- **Separate endpoints** — `/controlhub/*` vs `/api/v1/*`

### 2. Use ControlHub Token

```bash
GET /api/v1/controlhub/tenants
Authorization: Bearer <controlhub_access_token>

# Returns all tenants (company admin view)
```

Protected ControlHub routes require both:
1. Valid JWT (JWTMiddleware)
2. ControlHub role + company_admin level (ControlHubOnly middleware)

### 3. ControlHub Endpoints (Protected)

| Endpoint | Method | Purpose |
|---|---|---|
| `/controlhub/auth/login` | POST | Login as company admin |
| `/controlhub/auth/refresh` | POST | Refresh access token |
| `/controlhub/auth/logout` | POST | Logout (clear refresh token) |
| `/api/v1/controlhub/tenants` | GET | List all tenants |
| `/api/v1/controlhub/tenants` | POST | Create new tenant |

---

## File Upload (Optional)

The template supports pluggable file storage via the `FileUploader` interface.

### Implementing Local Storage

```go
// services/local_uploader.go
type LocalUploader struct {
    basePath string
}

func (u *LocalUploader) Upload(bucket, name string, data []byte) (string, error) {
    path := filepath.Join(u.basePath, bucket, name)
    return path, os.WriteFile(path, data, 0644)
}
```

### Implementing MinIO

```go
// services/minio_uploader.go
type MinIOUploader struct {
    client *minio.Client
}

func (u *MinIOUploader) Upload(bucket, name string, data []byte) (string, error) {
    info, err := u.client.PutObject(context.Background(), bucket, name, 
        bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{})
    return u.client.GetObjectURL(bucket, name)
}
```

In `main.go`:

```go
var uploadService global.FileUploader

if cfg.FileStorageType == "minio" {
    uploadService = NewMinIOUploader(cfg)
} else {
    uploadService = NewLocalUploader("/tmp/uploads")
}

app := global.New(cfg, sugar, emailService, uploadService)
```

---

## Email (Optional)

The template supports pluggable email services via the `EmailSender` interface.

### Implementing SMTP

```go
// services/smtp_sender.go
type SMTPSender struct {
    host string
    port string
    user string
    pass string
}

func (s *SMTPSender) Send(to, subject, body string) error {
    msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", 
        s.user, to, subject, body)
    return smtp.SendMail(fmt.Sprintf("%s:%s", s.host, s.port), 
        smtp.PlainAuth("", s.user, s.pass, s.host), s.user, []string{to}, []byte(msg))
}
```

In `main.go`:

```go
emailService := services.NewSMTPSender(cfg)
```

---

## Logging

The template uses **Zap** for structured logging.

```go
// In any handler or service
logger := app.Logger

logger.Info("User logged in")
logger.Warnf("Request timeout: %s", endpoint)
logger.Errorf("Database error: %v", err)
logger.Debugf("Query: %s", query)
```

Logs include:
- Structured fields (tenant_id, user_id, etc.)
- Stack traces for errors
- Request/response timings
- All state-changing operations (via AuditMiddleware)

---

## Testing

### Unit tests

```bash
go test ./services/...
go test ./handlers/...
```

### Integration tests

```bash
# Requires Docker + PostgreSQL running
go test -tags=integration ./tests/...
```

---

## Production Deployment

### Pre-flight Checklist

- [ ] JWT_SECRET is a strong random value (32+ chars)
- [ ] Database credentials use separate app/migration users
- [ ] CORS_ALLOWED_ORIGINS is restricted to your frontend domain
- [ ] DB connection pool is tuned for your load (DBPoolMaxOpen, DBPoolMaxIdle)
- [ ] Logs are sent to centralized logging system (ELK, Datadog, etc.)
- [ ] Audit logs are retained per compliance requirements
- [ ] DB backups are automated

### Deployment Steps

1. Set environment variables in production
2. Run migrations: `go run main.go` (only needs to succeed once)
3. Deploy binary to app server
4. Monitor `/api/v1/health` for readiness

### Scaling

**Horizontal:**
- Backend is stateless
- Multiple instances share the same DB
- Load balance traffic across instances

**Vertical:**
- Increase DBPoolMaxOpen if DB queries are queued
- Increase APP_PORT to handle more connections

---

## Common Tasks

### Add a permission

```go
permission := models.Permission{
    Code:     "documents.archive",
    TenantID: tenantID,
    Category: "documents",
    Description: "Archive documents",
}
db.Create(&permission)
```

### Grant permission to role

```go
db.Model(&role).Association("Permissions").Append(&permission)
```

### Create a custom role

```go
role := models.Role{
    TenantID:    tenantID,
    Name:        "Reviewer",
    Slug:        "reviewer",
    Description: "Can review documents",
}
db.Create(&role)

// Add specific permissions
db.Model(&role).Association("Permissions").Append([]uuid.UUID{perm1.ID, perm2.ID})
```

### View audit logs

```bash
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/audit-logs?resource=documents&action=DELETE
```

---

## Troubleshooting

### "Tenant not found"
- Check X-Tenant-ID header or JWT tenant_id claim
- Verify tenant exists in database: `SELECT * FROM tenants WHERE id = '...';`

### "Insufficient permissions"
- Verify user's role has required permission
- Check for conflicting permission overrides
- Query: `SELECT * FROM user_permission_overrides WHERE user_id = '...';`

### Database connection timeout
- Check DB_HOST and DB_PORT
- Verify firewall allows connections
- Check connection pool: `SELECT count(*) FROM pg_stat_activity;`

### JWT validation fails
- Ensure JWT_SECRET matches value used to sign token
- Check token expiry: `jq -R 'split(".")[1] | @base64d | fromjson' <<< $TOKEN`

---

## Directory Structure

```
.
├── main.go                 # Entry point, wiring, startup
├── go.mod                  # Dependencies
├── .env.example            # Configuration template
├── docker-compose.yml      # PostgreSQL
├── Makefile                # Build/run commands (optional)
│
├── db/                     # Database layer
│   ├── db.go               # Connection manager, singleton
│   ├── migrations.go       # Migration executor
│   └── migrations/         # SQL migration files
│       ├── 001_core_schema.sql
│       ├── 002_rls_policies.sql
│       └── 003_example_projects_feature.sql
│
├── models/                 # Domain entities
│   ├── base.go             # BaseModel, TenantBaseModel
│   ├── user.go             # User, Role, Permission, Override
│   ├── audit.go            # AuditLog
│   └── meta.go             # Tenant
│
├── middleware/             # Request processing pipeline
│   ├── auth.go             # JWT validation
│   ├── tenant.go           # Tenant resolution & DB scoping
│   ├── rbac.go             # Permission checking
│   ├── audit.go            # State-change logging
│   └── cors.go             # CORS headers
│
├── modules/                # Feature-specific handler/service pairs
│   ├── auth/               # Authentication module
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── models.go
│   │   └── public.go
│   ├── users/              # User management
│   ├── rbac/               # RBAC management
│   └── _template/          # Template for new modules
│       ├── handler.go
│       ├── service.go
│       ├── models.go
│       └── public.go
│
├── routes/                 # Route registration
│   └── routes.go           # Grouped endpoint registration
│
├── global/                 # Shared infrastructure
│   └── app.go              # App struct (Config, Logger, Email, Upload)
│
├── pkg/                    # Shared packages
│   ├── config.go           # Config loading & validation
│   └── response/           # Response helpers
│       └── response.go
│
└── docs/                   # Documentation
    ├── API.md              # API reference
    └── DEPLOYMENT.md       # Deployment guide
```

---

## License

[Your License Here]

---

## Support

Questions or issues? Open an issue on GitHub.
