# Avec Go Backend Template

**A production-ready, reusable backend starter built from the UNIDIMS codebase.**

---

## What You're Getting

This template extracts the best patterns from the UNIDIMS backend and packages them as a clean starting point for new Go/Echo/PostgreSQL projects. It's not UNIDIMS with features removed — it's a deliberate **distillation of infrastructure + patterns + architecture**, with everything DIMS-specific left behind.

### Core Strength: Built-In Multi-Tenancy

Unlike generic Go templates, this one has multi-tenancy baked in:
- Middleware-driven tenant scoping (WHERE clauses automatic)
- Single-tenant mode fallback (set `MULTI_TENANT_ENABLED=false`)
- Simple DB setup: the template uses a single database role by default (no RLS or separate migration-only role). Projects that need DB-level enforcement can add RLS and a migration role later.

### What's Inside

```
avec-go-template/
├── db/                    # Connection manager + migrations
├── models/                # Base, user/RBAC, audit schemas
├── middleware/            # JWT, tenant, RBAC, audit, CORS
├── modules/               # Feature scaffolds (auth, users, rbac, _template)
├── services/              # Infrastructure services (email, upload)
├── global/                # Shared App infrastructure
├── pkg/                   # Config + response helpers
├── routes/                # Route registration
├── main.go                # Clean wiring
├── Makefile               # Build commands
├── docker-compose.yml     # PostgreSQL setup
├── .env.example           # Configuration template
├── go.mod                 # Minimal dependencies (6 core)
└── docs/
    ├── README.md          # Full documentation
    └── TEMPLATE_GUIDE.md  # How to use and extend
```

---

## Philosophy

### Minimal, Not Maximal

The template includes:
- ✅ Authentication (JWT)
- ✅ RBAC (roles + permissions + per-user overrides)
- ✅ Audit logging
- ✅ Multi-tenancy infrastructure
- ✅ Structured logging (Zap)
- ✅ Standard response format

It **explicitly excludes**:
- ❌ WebSockets
- ❌ Background jobs
- ❌ Caching
- ❌ File upload (only interfaces defined)
- ❌ Email templates (only infrastructure)
- ❌ Pagination helpers
- ❌ OpenTelemetry
- ❌ API documentation generators

**Reasoning:** Every excluded feature has a legitimate reason to exist in specific projects. None has a reason to exist in *every* project. The template's value is in what it deliberately omits as much as in what it provides.

### Three Clear Architectural Patterns

1. **New Modules** (handler + service + DTOs) — Use for domain features
2. **Thin Wrappers** (delegates to legacy) — Use during gradual migration
3. **Infrastructure** (interfaces in global.App) — Use for cross-cutting concerns

This clarity prevents the pattern confusion that derails codebases over time.

### Security-First Design

- Tenant scoping enforced by middleware and query-level filters
- Single-role database by default (no RLS enabled) — optional RLS/multi-role setups can be added per-project if required by your threat model
- Statement timeout prevents query runaway
- Audit middleware logs all state changes
- Permission overrides tracked with reasoning

Not overthinking. Not under-secured.

---

## First 15 Minutes

```bash
# 1. Clone and rename
git clone <repo> my-project
cd my-project
rm -rf .git && git init

# 2. Update go.mod
sed -i 's/backend/github.com\/yourorg\/my-project\/backend/g' go.mod

# 3. Configure
cp .env.example .env
# Edit .env: DB_HOST, DB_NAME, JWT_SECRET

# 4. Start
make db-up
go run main.go

# 5. Verify
curl http://localhost:8080/api/v1/health
```

That's it. You have a working auth backend with RBAC and multi-tenancy.

---

## Reusable Components

These are extracted directly from DIMS and work in any project:

| Component | Reuse Strategy |
|---|---|
| **BaseModel** | Embed in your entities (UUID PK + timestamps + soft delete) |
| **TenantBaseModel** | Embed for tenant-scoped entities |
| **DB connection manager (single-role)** | Copy `db.go` and the migrations approach; add a migration-only role / RLS later if desired |
| **TenantMiddleware** | Scopes queries via WHERE clause |
| **RBACMiddleware** | Role + permission override checking |
| **AuditMiddleware** | Automatic state-change logging |
| **Response helpers** | Consistent API response format |
| **Module scaffold** | Copy _template/ for new features |

---

## What Was Intentionally Removed from DIMS

To make this a template, not a modified DIMS fork:

1. **Domain models** (Document, Department, Workflow, Signature, etc.) → Not included. You build yours.

2. **25 migration files** → Replaced with 2 canonical migrations (core schema, example feature). RLS policies were intentionally omitted from the template to keep onboarding simple; add them per-project if needed.

3. **62 config variables** → Reduced to 15 required. Domain-specific config (AI providers, Google OAuth) removed.

4. **Hardcoded CORS** → Made config-driven (`CORS_ALLOWED_ORIGINS` env var).

5. **25 legacy handlers/services** → Not included. They exemplify the old pattern; use modules/ instead.

6. **DIMS-specific dependencies** → Removed: excelize (Excel), fpdf (PDF), google.golang.org/api, AI clients.

7. **Shim modules** (documents, workflows, etc.) → Not included. They're DIMS migration artifacts.

8. **Service constructors receiving concrete DB** → Changed to receive `*global.App` + per-method `*gorm.DB`.

---

## Design Decisions Explained

### Q: Why embed models instead of library inheritance?
**A:** GORM's field reflection doesn't work well with library-provided base types. You need to own your `BaseModel` to customize hooks, add fields, or override behavior. Copying is safer than inheriting.

### Q: Why a single database role by default?
**A:** The template uses a single DB role by default to reduce configuration and simplify onboarding. Two-role and RLS strategies are intentionally omitted from the template; projects that require DB-level enforcement can add them later.

### Q: Why Session(NewDB:false) in tenant middleware?
**A:** Ensures WHERE clause persists through Preload(). Without it, preloaded associations can escape tenant boundaries. It's a safety pattern worth encoding.

### Q: Why no RLS by default?
**A:** RLS adds operational complexity (additional roles, connection management, and migration privileges). For a reusable starter we prefer a simpler default (tenant scoping via middleware). Teams with stricter DB-level requirements can add RLS and a migration role in their project-specific setup.

### Q: Why interfaces for Email and Upload in global.App?
**A:** Because projects have wildly different needs. SMTP vs SendGrid. Local filesystem vs MinIO vs S3. By defining interfaces, you swap implementations without changing module code.

### Q: Why no background jobs in the template?
**A:** Job systems are architecturally invasive (need workers, queues, error handling). No generic choice works everywhere. Projects that need jobs add gocron or Temporal; projects that don't don't waste the complexity.

---

## What to Extend

### Add a new module

```bash
make new-module NAME=invoices
```

Implement service/handler/models. Register routes. Done.

### Add infrastructure service

Define interface in `global/app.go`. Implement concrete type. Inject in `main.go`. Use in modules.

Example: Cache, queue, auth provider, email service, file storage.

### Add feature migrations

Create SQL file in `db/migrations/`. RLS policies are optional — include them in feature migrations only if your project requires DB-level row isolation. Migrations run automatically at startup.

---

## What Not to Do

- ❌ Store `*gorm.DB` in service structs (pass as method parameter)
- ❌ Couple modules to each other (use dependency injection)
- ❌ Bypass middleware (don't create routes outside the api group)
- ❌ Hardcode tenant_id (middleware handles it)
Note: RLS is optional in this template — consider it for defense-in-depth when your project's threat model requires it.

---

## Next Steps

1. **Read docs/README.md** — Full API reference and configuration
2. **Read docs/TEMPLATE_GUIDE.md** — How to extend the template
3. **Run `make help`** — See available commands
4. **Clone the template** — Start your project
5. **Add your first module** — See Phase 2 in TEMPLATE_GUIDE.md

---

## Maintenance

This template is a snapshot of DIMS patterns at a point in time. When DIMS improvements happen, you manually port back if they're generic infrastructure improvements (not DIMS-specific).

Examples of worth-porting-back:
- Improved RLS patterns
- Better error handling in middleware
- Optimized DB connection pooling
- Security fixes

Examples of not-worth-porting-back:
- New DIMS features (Document AI, Workflows, etc.)
- DIMS-specific models or handlers
- Domain-specific RBAC logic

---

## Questions?

See docs/TEMPLATE_GUIDE.md for detailed walkthrough.

Open an issue or reach out to the team.

**Happy building.**
