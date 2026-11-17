
# Year Planner SaaS

A full-year planning and habit-tracking platform
with personal and collaborative workspaces.

## 1. Overview

Year Planner allows individuals and teams to:

- Visualize an entire year.
- Plan weekly activities.
- Organize daily schedules.
- Track habits and streaks.
- Manage quarterly goals.
- Track time and activities.
- Collaborate in workspaces.
- Assign granular permissions.

## 2. Technology

Frontend:
- React
- TypeScript
- Vite
- Tailwind CSS
- shadcn/ui
- TanStack Query

Backend:
- Go
- net/http
- Chi
- MongoDB Go Driver

Database:
- MongoDB Atlas

API:
- REST
- OpenAPI

Testing:
- Go testing
- Vitest
- Playwright

Deployment:
- Docker
- Compiled Go binary

## 3. Repository Structure

year-planner/

  AGENTS.md
  INTENT.md
  SCOPE.md
  SKILL.md
  ARCHITECTURE.md
  PLAN.md
  TASKS.md
  DECISIONS.md
  README.md

  web/
  api/
  docs/
  scripts/

## 4. Getting Started

### Requirements

- Go stable toolchain.
- Node.js LTS for frontend tooling.
- MongoDB Atlas account or local MongoDB.
- Git.
- Docker (recommended).

### Backend Setup

Navigate to:

```powershell
cd api
if (-not (Test-Path .env)) { Copy-Item .env.example .env }
```

Install dependencies:

```powershell
go mod tidy
```

Set the values in `api/.env` for your local MongoDB instance.
Use a URL-encoded password in `MONGODB_URI` when it contains
reserved URI characters. `MONGODB_DATABASE` selects the app database.

Never commit `.env` or use the MongoDB root account in production.
Create a dedicated database user with only the permissions the app needs.
Password recovery requires `SMTP_HOST`, `SMTP_PORT`, and `SMTP_FROM`;
SMTP credentials are optional only for unauthenticated local relays.
Without SMTP configured, recovery endpoints return an explicit unavailable
response. Production requires HTTPS for `APP_BASE_URL` and TLS-capable SMTP
(implicit TLS on port 465 or STARTTLS). Reset emails are stored in an encrypted
MongoDB outbox and retried; keep `SESSION_SECRET` stable while queued emails
may still be pending. Production session secrets must be at least 32 bytes.

Start backend:

```powershell
go run ./cmd/server
```

The server checks MongoDB connectivity and creates required indexes before
listening. Liveness and readiness endpoints are available at
`/health/live` and `/health/ready`.
Interactive API documentation is at `http://localhost:8080/docs`.

### Frontend Setup

From the repository root:

```powershell
cd web
if (-not (Test-Path .env)) { Copy-Item .env.example .env }
```

Install dependencies:

```powershell
npm install
```

Start the frontend:

```powershell
npm run dev
```

The Vite development server proxies `/api` requests to the Go API on
`http://localhost:8080`. Start the backend first. The frontend currently
provides registration, sign-in, and a year-planner preview; annual plan data
will be connected in a later phase.

### Build the API container

From the repository root, build with:

```powershell
docker build -t yearplanner-api ./api
```

The image runs as a non-root user. Provide MongoDB and session configuration
through the runtime environment; secrets are not included in the image.

## 5. Development Commands

Backend:

gofmt -w .
go vet ./...
go test ./...
go test -race ./...

Frontend:

npm run lint
npm run typecheck
npm run test
npm run build

Run only commands that are configured in the
actual repository.

## 6. API Documentation

OpenAPI specification:

api/openapi/

All API endpoints use:

/api/v1

## 7. Documentation Reading Order

New developers and AI agents must read:

1. AGENTS.md
2. INTENT.md
3. SCOPE.md
4. SKILL.md
5. ARCHITECTURE.md
6. PLAN.md
7. TASKS.md
8. DECISIONS.md

## 8. Project Progress

Current phase:
Phase 1 — Authentication

Current implementation status:
In progress; see `TASKS.md` for verified work and remaining phases.

The authoritative task status is TASKS.md.

## 9. Development Rules

- Follow AGENTS.md.
- Preserve the supplied UI.
- Use Go for the backend.
- Use MongoDB for persistence.
- Enforce backend authorization.
- Maintain API documentation.
- Write tests for business logic.
- Update TASKS.md after verification.
- Record architecture changes in DECISIONS.md.

## 10. Production Principles

Never deploy without:

- Secure authentication.
- Tenant isolation.
- RBAC tests.
- Database backups.
- Monitoring.
- Environment secret management.
- Production build verification.
- Deployment and rollback documentation.

## 11. License

License and distribution terms must be decided
by the project owner before public distribution.
