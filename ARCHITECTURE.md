
# ARCHITECTURE.md

# System Architecture

## 1. Architecture Overview

Architecture: Modular Monolith

Backend: Go
Database: MongoDB
Frontend: React SPA

The React application communicates with the Go API
through versioned REST endpoints.

The Go API handles:
- Authentication.
- Authorization.
- Business rules.
- Calendar calculations.
- Database access.
- Aggregations.
- Audit events.

The React application handles:
- UI rendering.
- User interactions.
- Navigation.
- Local presentation state.
- Server-state caching.

## 2. High-Level Architecture

Browser
  |
  v
React Application
  |
  v
API Client
  |
  v
Go HTTP Server
  |
  +-- Authentication
  |
  +-- Tenant Resolution
  |
  +-- Authorization
  |
  +-- Domain Services
  |
  +-- MongoDB Repositories
  |
  v
MongoDB Atlas

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
    src/
      app/
      components/
      features/
      hooks/
      lib/
      routes/
      types/

  api/
    cmd/
      server/
        main.go

    internal/
      config/
      http/
        middleware/
        handlers/
        response/
        routes/

      auth/
      users/
      workspaces/
      memberships/
      permissions/

      planner/
      weeks/
      days/
      habits/
      goals/
      activities/
      analytics/

      audit/
      store/

    openapi/
    migrations/
    go.mod
    Dockerfile

  docs/
    api/
    database/
    security/
    testing/
    deployment/

  scripts/

## 4. Backend Layering

### HTTP Handler

Responsibilities:
- Parse path parameters.
- Decode JSON.
- Validate request shape.
- Resolve authenticated actor.
- Call service.
- Write HTTP response.

Handlers must not contain complex business rules.

### Service

Responsibilities:
- Business logic.
- Permission orchestration.
- Ownership validation.
- Domain validation.
- Transaction coordination.
- Cross-module operations.

### Repository

Responsibilities:
- MongoDB queries.
- Inserts.
- Updates.
- Deletes.
- Aggregations.
- Index-compatible query design.

Repositories must always apply tenant scope.

### Domain Models

Responsibilities:
- Domain structures.
- Enumerations.
- Business invariants.
- Domain-level validation.

Avoid importing HTTP-specific packages
into domain models.

## 5. Frontend Architecture

The supplied React UI is preserved when present. This repository initially
contained no frontend files; the project owner approved a new initial UI.

Suggested feature organization:

web/src/

  app/
    providers/
    router/

  components/
    layout/
    navigation/
    ui/

  features/
    auth/
    workspace/
    year-board/
    week-drawer/
    day-plans/
    habits/
    goals/
    activities/
    members/
    permissions/

  lib/
    api/
    query/
    dates/
    permissions/

  types/

Each feature owns its:
- Components.
- API functions.
- Hooks.
- Types.
- Tests.

Avoid one giant dashboard component.

### Initial UI Mapping

| UI | Frontend location | API/data status |
| --- | --- | --- |
| Registration and sign-in | `web/src/features/auth/` | Connected to the Phase 1 auth endpoints |
| Authenticated year-board shell | `web/src/features/year-board/` | Preview only; annual plan API is not implemented |
| Typed API client and query provider | `web/src/lib/`, `web/src/main.tsx` | Auth API only |

The year-board preview must not be presented as persisted planner data.
Connect it to the annual plan and calendar APIs after their backend phase is
complete.

## 6. Multi-Tenant Ownership

Two resource scopes exist.

### Personal

Resource:
- ownerUserId.
- scope.kind = personal.

Authorization:
Authenticated user must match ownerUserId.

### Workspace

Resource:
- workspaceId.
- scope.kind = workspace.

Authorization:
Authenticated user must have active membership
and the required permission.

Personal resources never become shared
merely because a user joins a workspace.

## 7. MongoDB Collections

### users

Fields:
- _id
- emailNormalized
- displayName
- avatarUrl
- status
- createdAt
- updatedAt
- lastLoginAt

Index:
- Unique emailNormalized.

### sessions

Fields:
- _id
- userId
- tokenHash
- expiresAt
- createdAt
- revokedAt
- lastUsedAt

Indexes:
- Unique tokenHash.
- userId.
- Expiration cleanup strategy.

Never store raw session secrets.

### workspaces

Fields:
- _id
- name
- slug
- ownerUserId
- settings
- schemaVersion
- createdAt
- updatedAt
- deletedAt

Indexes:
- Globally unique slug, including soft-deleted workspaces.
- ownerUserId + deletedAt + createdAt + _id for owner-scoped listing.

Soft-deleting a workspace does not release its slug, so a prior workspace
identifier can never be reassigned to a different tenant.

`settings` is a typed object with no fields until workspace-level setting
keys are approved in product scope. Do not accept arbitrary client-supplied
settings maps.

### memberships

Fields:
- _id
- workspaceId
- userId
- roleKey
- grants
- denies
- status
- invitedBy
- acceptedAt
- createdAt
- updatedAt

Index:
Unique workspaceId + userId.

Membership status values are `active` and `removed`. Removed membership
records are retained but must not grant access; pending invitation state stays
in the invitations collection. `roleKey`, `grants`, and `denies` remain
strings until the Phase 3 permission catalogue is defined.

### invitations

Fields:
- _id
- workspaceId
- emailNormalized
- tokenHash
- roleKey
- invitedBy
- expiresAt
- acceptedAt
- revokedAt

Indexes:
- Unique tokenHash.
- workspaceId + emailNormalized.
- Expiration cleanup strategy.

### annualPlans

Fields:
- _id
- scope
- year
- title
- timezone
- weekStartsOn
- visibility
- createdBy
- createdAt
- updatedAt
- archivedAt

Indexes:
- Personal owner + year.
- Workspace + year.

Define active-plan uniqueness explicitly.

### weeks

Fields:
- _id
- planId
- scope
- weekYear
- weekNumber
- startDate
- endDate
- focus
- notes
- createdBy
- createdAt
- updatedAt

Index:
Unique planId + weekYear + weekNumber.

### dayPlans

Fields:
- _id
- planId
- scope
- localDate
- items
- createdAt
- updatedAt

Index:
Unique planId + localDate.

### habits

Fields:
- _id
- planId
- scope
- name
- description
- category
- color
- schedule
- targetPerWeek
- startDate
- endDate
- archivedAt
- createdAt
- updatedAt

### habitEntries

Fields:
- _id
- habitId
- planId
- scope
- localDate
- status
- value
- note
- completedBy
- createdAt
- updatedAt

Index:
Unique habitId + localDate.

### goals

Fields:
- _id
- planId
- scope
- title
- description
- quarter
- status
- priority
- dueDate
- completedAt
- createdAt
- updatedAt

### activities

Fields:
- _id
- planId
- scope
- title
- category
- startAt
- endAt
- durationMinutes
- localDate
- source
- goalId
- habitId
- createdAt
- updatedAt

### auditEvents

Fields:
- _id
- workspaceId
- actorUserId
- action
- targetType
- targetId
- metadata
- occurredAt

Audit records must be append-only
for ordinary application users.

## 8. Database Design Rules

Do not embed an entire year into one document.

Do not embed unlimited activity history.

Do not embed unlimited habit entries.

Use separate collections for growing data.

Add indexes based on actual query patterns.

Use projections for summary endpoints.

Use optimistic concurrency or version fields
where concurrent editing requires conflict detection.

## 9. Calendar Model

Use ISO week numbering.

A week is:

Monday 00:00 through Sunday 23:59
in the plan's calendar timezone.

Persist:
- ISO week-year.
- ISO week number.
- Local date identity.

A week can overlap two calendar years.

The same ISO week must have the same identity
when accessed from either adjacent annual view.

The annual board displays all ISO weeks
that intersect the selected calendar year.

A year can have 52 or 53 ISO weeks.

## 10. API Design

Prefix:

/api/v1

Authentication:
GET    /auth/session
POST   /auth/register
POST   /auth/login
POST   /auth/logout

Users:
GET    /users/me
PATCH  /users/me

Workspaces:
POST   /workspaces
GET    /workspaces
GET    /workspaces/:workspaceId
PATCH  /workspaces/:workspaceId
DELETE /workspaces/:workspaceId

The workspace PATCH endpoint currently supports renaming only. Slug and owner
are not client-updatable, and settings remain immutable until their fields are
defined in product scope.

Members:
GET    /workspaces/:workspaceId/members
PATCH  /workspaces/:workspaceId/members/:memberId
DELETE /workspaces/:workspaceId/members/:memberId

Invitations:
POST   /workspaces/:workspaceId/invitations
POST   /invitations/:token/accept
POST   /invitations/:token/reject

Plans:
POST   /plans
GET    /plans
GET    /plans/:planId
PATCH  /plans/:planId
DELETE /plans/:planId

Year Board:
GET /plans/:planId/year-board?year=2026

Weeks:
GET   /plans/:planId/weeks/:weekYear/:weekNumber
PATCH /plans/:planId/weeks/:weekYear/:weekNumber

Days:
GET /plans/:planId/days?from=&to=
PUT /plans/:planId/days/:localDate

Habits:
GET    /plans/:planId/habits
POST   /plans/:planId/habits
PATCH  /plans/:planId/habits/:habitId
DELETE /plans/:planId/habits/:habitId

Habit Entries:
PUT /plans/:planId/habits/:habitId/entries/:localDate
GET /plans/:planId/habits/:habitId/entries

Goals:
GET    /plans/:planId/goals
POST   /plans/:planId/goals
PATCH  /plans/:planId/goals/:goalId
DELETE /plans/:planId/goals/:goalId

Activities:
GET    /plans/:planId/activities
POST   /plans/:planId/activities
PATCH  /plans/:planId/activities/:activityId
DELETE /plans/:planId/activities/:activityId

Analytics:
GET /plans/:planId/analytics/summary

Audit:
GET /workspaces/:workspaceId/audit-events

## 11. Year Board Response

The year-board endpoint must return a compact,
purpose-built response.

It should include:

- Plan information.
- Year information.
- Week metadata.
- Current week identity.
- Activity summaries.
- Habit summaries.
- Weekly completion counts.
- Quarterly goal summaries.
- Overall progress.

Avoid sending full activity histories.

The UI should not need 53 individual requests.

## 12. API Response Format

Success:

{
  "data": {},
  "meta": {
    "requestId": "..."
  }
}

Error:

{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request"
  },
  "meta": {
    "requestId": "..."
  }
}

## 13. Authorization Pipeline

Every protected request follows:

1. Request arrives.
2. Request ID middleware.
3. Security and recovery middleware.
4. Session authentication.
5. User status validation.
6. Resource resolution.
7. Tenant membership verification.
8. Permission resolution.
9. Resource-specific ownership checks.
10. Service execution.
11. Repository execution with tenant filters.
12. Response mapping.
13. Audit event when applicable.

Never skip this pipeline.

## 14. Year Board Data Flow

User opens selected year.

React:
- Reads planId and year from URL.
- Calls year-board endpoint.
- TanStack Query caches response.
- Renders year grid.
- Renders sidebar summaries.

User selects week.

React:
- Updates selected week URL state.
- Opens week drawer.
- Loads week detail if not already available.

User completes habit.

React:
- Optimistically updates the habit entry.
- Calls Go API.
- Go validates session and permission.
- Go validates habit ownership.
- Go writes MongoDB entry.
- Go returns authoritative result.
- TanStack Query reconciles relevant summaries.

## 15. Production Deployment

Frontend:
Static React build.

Backend:
Compiled Go binary.

Database:
MongoDB Atlas.

Deployment:
Docker-compatible hosting.

Use:
- Environment configuration.
- Secret management.
- TLS.
- Health endpoints.
- Backups.
- Structured logs.
- Monitoring.
- Graceful shutdown.

No production secrets in source control.
