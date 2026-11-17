
# SKILL.md

# Year Planner SaaS — Engineering Skill

## 1. Engineering Objective

Implement a production-grade multi-tenant planner.

Priorities:

1. Security.
2. Data correctness.
3. Calendar correctness.
4. Maintainability.
5. Performance.
6. User experience.
7. Development speed.

Never sacrifice tenant isolation for convenience.

## 2. Technology Stack

### Backend

- Go stable release.
- net/http.
- Chi router.
- Official MongoDB Go Driver.
- MongoDB Atlas.
- slog structured logging.
- OpenAPI.
- Context-aware database operations.

### Frontend

The user supplies the UI.

Use:
- React.
- TypeScript.
- Vite.
- Tailwind CSS.
- shadcn/ui.
- TanStack Query.
- React Hook Form.
- Zod.

Do not replace the user's supplied design.

### Testing

Backend:
- Go testing.
- httptest.
- Integration tests.
- Race detector.

Frontend:
- Vitest.
- React Testing Library.
- Playwright.

## 3. Go Backend Standards

### Project Layout

api/
  cmd/server/main.go
  internal/config/
  internal/http/
  internal/auth/
  internal/users/
  internal/workspaces/
  internal/memberships/
  internal/permissions/
  internal/planner/
  internal/habits/
  internal/goals/
  internal/activities/
  internal/analytics/
  internal/audit/
  internal/store/
  openapi/
  go.mod

### Coding Standards

- Use idiomatic Go.
- Run gofmt.
- Run go vet.
- Avoid unnecessary interfaces.
- Avoid global mutable state.
- Use explicit dependencies.
- Pass context.Context.
- Keep handlers small.
- Keep business logic in services.
- Keep persistence in repositories.
- Return typed domain errors.
- Avoid panic for normal application errors.
- Avoid unnecessary reflection.

### HTTP Server

Configure:
- ReadHeaderTimeout.
- ReadTimeout.
- WriteTimeout.
- IdleTimeout.
- Request body limits.
- Graceful shutdown.
- Request ID middleware.
- Recovery middleware.
- Authentication middleware.
- Authorization middleware.

Use a lightweight router.

Do not build a custom framework.

### MongoDB

- One reusable MongoDB client per process.
- Configure connection pooling.
- Configure operation timeouts.
- Use context-aware operations.
- Create indexes during controlled startup/migration.
- Use projections.
- Avoid unbounded arrays.
- Use compound indexes for tenant-scoped queries.
- Handle duplicate key errors.
- Handle transient database errors.
- Use transactions only when required.

## 4. MongoDB Document Standards

Use:
- ObjectId references.
- Explicit BSON tags.
- Explicit JSON tags.
- UTC timestamps.
- Date-only strings for local calendar dates.
- schemaVersion for documents requiring evolution.

Never expose MongoDB internal fields unnecessarily.

Never accept a persistence model directly
as a client update payload.

Map approved fields explicitly.

## 5. Multi-Tenant Security

Every tenant-owned document must identify
its ownership scope.

Personal resource:

scope.kind = personal
ownerUserId = authenticated user

Workspace resource:

scope.kind = workspace
workspaceId = authorized workspace

Every repository method must receive enough
trusted context to enforce the ownership boundary.

Never fetch tenant-owned documents by ID alone.

Never trust workspaceId from a client without
checking the authenticated user's membership.

## 6. RBAC

Permissions are stable strings.

Examples:

workspace.read
workspace.update
workspace.delete

members.read
members.invite
members.remove
members.changeRole

planner.read
planner.create
planner.update
planner.delete

habits.read
habits.create
habits.update
habits.complete

goals.read
goals.create
goals.update
goals.delete

activities.read
activities.create
activities.update
activities.delete

analytics.read
audit.read

Roles are bundles of permissions.

Default roles:

Owner:
Full workspace control.

Admin:
Workspace administration except owner-only actions.

Editor:
Create and update shared planner content.

Member:
Read shared content and make permitted contributions.

Viewer:
Read-only access.

Custom:
Explicitly configured permission set.

### Permission Resolution

Recommended model:

1. Resolve base role permissions.
2. Add explicitly granted permissions.
3. Apply explicit denials.
4. Denial takes precedence.
5. Check resource-specific ownership restrictions.
6. Check actor's delegation authority.

A user must not grant permissions they are
not authorized to delegate.

Authorization must happen on the backend.

## 7. Authentication

Use secure server-managed sessions.

Cookie requirements:

- HttpOnly.
- Secure in production.
- SameSite policy.
- Appropriate Path.
- Appropriate expiration.

Protect cookie-authenticated mutations
against CSRF.

Use Argon2id if implementing passwords.

Do not invent cryptographic primitives.

Rate-limit authentication endpoints.

Support session revocation.

Do not store session secrets in frontend storage.

## 8. REST API Standards

Prefix:

/api/v1

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
    "code": "FORBIDDEN",
    "message": "Permission denied"
  },
  "meta": {
    "requestId": "..."
  }
}

Use stable error codes.

Use correct HTTP status codes.

Validate all input.

Document every endpoint in OpenAPI.

Avoid leaking internal errors.

## 9. API Contract

The Go API owns business rules.

The frontend owns presentation.

Use OpenAPI as the contract between both systems.

Frontend API functions should be typed.

Do not duplicate business rules in React.

If an API contract changes:
1. Update OpenAPI.
2. Update frontend DTOs.
3. Update integration tests.
4. Document breaking changes.

## 10. TanStack Query

Use centralized query keys.

Example:

planKeys.all()
planKeys.detail(planId)
planKeys.yearBoard(planId, year)
planKeys.week(planId, weekYear, weekNumber)
habitKeys.entries(habitId, from, to)

Use typed query functions.

Avoid fetching individual weeks separately
when the annual board can be loaded as one summary.

Use optimistic updates only for safe operations.

On optimistic update:
- Cancel relevant requests.
- Snapshot old data.
- Update cache.
- Roll back on failure.
- Reconcile on success.

Do not duplicate server state unnecessarily.

## 11. Calendar Standards

Use ISO week numbering.

Weeks run Monday through Sunday.

Support:
- 52 weeks.
- 53 weeks.
- Leap years.
- Year crossover.
- ISO week-year.

Store:
- weekYear.
- weekNumber.
- localDate.
- plan timezone.

Use UTC for actual timestamps.

Use YYYY-MM-DD for local all-day date identity.

Never rely on implicit browser timezone parsing.

## 12. Performance

- Optimize the year-board summary endpoint.
- Use MongoDB projections.
- Add correct indexes.
- Avoid N+1 requests.
- Avoid unbounded MongoDB documents.
- Keep Go allocations reasonable.
- Reuse MongoDB client connections.
- Avoid unnecessary frontend rerenders.
- Lazy-load non-critical frontend modules.
- Measure before optimizing.

The annual board has only approximately
52 or 53 week cards.

Do not introduce virtualization without
a demonstrated performance problem.

## 13. Error Handling

Use typed errors for:

- Not found.
- Unauthorized.
- Forbidden.
- Validation failure.
- Conflict.
- Rate limited.
- Internal error.

Centralize HTTP error mapping.

Do not return raw database errors.

## 14. Logging and Observability

Use structured logs.

Include:
- Request ID.
- Route.
- Status.
- Latency.
- Actor ID where appropriate.
- Workspace ID where appropriate.

Never log:
- Passwords.
- Cookies.
- Session tokens.
- Invitation tokens.
- Private notes.
- Sensitive personal information.

Provide:
- Liveness endpoint.
- Readiness endpoint.
- Database health checks.
- Error monitoring.
- Slow query monitoring.

## 15. Security Testing

Test:
- IDOR attacks.
- Cross-workspace access.
- Cross-user personal data access.
- Privilege escalation.
- Expired invitation use.
- Invitation replay.
- Session revocation.
- Viewer mutation attempts.
- Admin owner-only actions.
- Invalid MongoDB query operators.
- Mass assignment.

## 16. Frontend Integration

Inspect the supplied UI first.

Create an integration map:

UI Component -> API Endpoint -> Query Key -> Mutation -> Database Resource

Preserve visual design.

Use accessible components.

Handle:
- Loading.
- Empty.
- Error.
- Success.
- Disabled.
- Unauthorized states.

## 17. Docker and Deployment

Build Go binary using a multi-stage Dockerfile.

Use a minimal runtime image.

Run as a non-root user.

Externalize configuration.

Never bake secrets into images.

Provide:
- Health checks.
- Graceful shutdown.
- Logging.
- Backup instructions.
- Rollback instructions.

## 18. Definition of Done

A feature is done only when:

- Backend implementation exists.
- Database persistence works.
- Authorization works.
- Frontend integration works where required.
- Tests pass.
- API documentation is updated.
- Error handling exists.
- Relevant acceptance criteria pass.

Never claim production readiness based
only on a successful compilation.
