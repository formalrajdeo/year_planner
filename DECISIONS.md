
# DECISIONS.md

# Architecture and Product Decision Register

## Decision Management Rules

Every important architectural or product decision
must be recorded here.

Use stable identifiers:

ADR-001
ADR-002
ADR-003

Statuses:

PROPOSED
ACCEPTED
REJECTED
SUPERSEDED

Never silently reverse an accepted decision.

If a decision must change:
- Create a new decision.
- Explain why.
- Explain the migration impact.
- Update architecture documentation.
- Update affected tasks.

---

## ADR-001 — Backend Language

Status: ACCEPTED

Decision:
Use Go as the application backend.

Reason:
- Lightweight runtime.
- Fast compiled executable.
- Strong concurrency support.
- Static typing.
- Standard library HTTP support.
- Straightforward container deployment.

Consequences:
- Backend code must use Go.
- Do not create Node.js API servers.
- Frontend tooling may use Node.js.

## ADR-002 — Database

Status: ACCEPTED

Decision:
Use MongoDB Atlas.

Reason:
- Flexible planner documents.
- Document-oriented domain model.
- Managed deployment.
- Compound indexes.
- Aggregation framework.

Consequences:
- Use official Go MongoDB Driver.
- Define schemas through typed models and validation.
- Avoid unbounded embedded arrays.
- Apply tenant-scoped indexes.

## ADR-003 — Frontend

Status: ACCEPTED

Decision:
Use the user's supplied React + TypeScript UI.

Technology:
- React.
- TypeScript.
- Vite.
- Tailwind.
- shadcn/ui.
- TanStack Query.

Consequences:
- Preserve supplied design.
- Integrate APIs into existing components.
- Do not generate a competing UI without approval.

## ADR-004 — Backend Architecture

Status: ACCEPTED

Decision:
Use a modular monolith.

Consequences:
- One Go API service.
- Modular domain packages.
- Shared MongoDB client.
- No premature microservices.

## ADR-005 — API Style

Status: ACCEPTED

Decision:
Use versioned REST APIs.

Prefix:
/api/v1

API contracts:
OpenAPI.

Consequences:
- Frontend consumes typed JSON APIs.
- Breaking changes require versioning or migration planning.

## ADR-006 — Multi-Tenancy

Status: ACCEPTED

Decision:
Support personal and workspace ownership scopes.

Personal:
Owned by an individual user.

Workspace:
Owned by a workspace.

Consequences:
- Personal resources remain private.
- Workspace membership does not imply personal-plan access.
- Every tenant-owned query must apply ownership filters.

## ADR-007 — Planning Model

Status: PROPOSED

Decision:
Support both shared workspace plans and individual plans.

Recommended behavior:
- Personal plan belongs to the user.
- Shared plan belongs to workspace.
- Individual workspace-context plan belongs to its designated user or explicitly defined workspace-member scope.

Open consideration:
Confirm whether individual workspace plans are strictly private
or visible to workspace administrators under an explicit policy.

## ADR-008 — Authorization

Status: ACCEPTED

Decision:
Use RBAC with granular permission strings.

Default roles:
- Owner.
- Admin.
- Editor.
- Member.
- Viewer.
- Custom.

Permission resolution:
Role grants + explicit grants - explicit denials.

Denials take precedence.

Consequences:
- Authorization enforced server-side.
- UI permission gates are presentation only.
- Privilege delegation must be restricted.

## ADR-009 — Calendar

Status: ACCEPTED

Decision:
Use ISO weeks.

Week:
Monday through Sunday.

Annual board:
Show all weeks intersecting the selected calendar year.

Consequences:
- Support 52 and 53 weeks.
- Store ISO week-year separately.
- Test year crossover.

## ADR-010 — Timezone

Status: ACCEPTED

Decision:
Timezone is configured per annual plan.

Default:
Use the user's preferred timezone when creating a plan.

Consequences:
- Store UTC timestamps.
- Preserve local date identity.
- Do not depend on browser timezone for business rules.

## ADR-011 — Authentication Sessions

Status: ACCEPTED

Decision:
Use secure server-managed sessions with HttpOnly cookies.

Consequences:
- Store session token hashes.
- Implement revocation.
- Protect against CSRF.
- Never expose session tokens to localStorage.

## ADR-012 — State Management

Status: ACCEPTED

Decision:
Use TanStack Query for server state.

Consequences:
- Centralized query keys.
- Typed API functions.
- Mutation invalidation.
- Carefully controlled optimistic updates.

## ADR-013 — Deployment

Status: ACCEPTED

Decision:
Build Go as a compiled production binary.

Frontend:
Static React assets.

Backend:
Docker-compatible Go service.

Consequences:
- Multi-stage Docker build.
- Non-root runtime.
- Externalized configuration.
- No secrets in images.

## ADR-014 — Redis

Status: PROPOSED

Decision:
Do not introduce Redis during initial development.

Reconsider only when measured requirements
justify distributed caching, rate limiting,
sessions, or asynchronous jobs.

## ADR-015 — Microservices

Status: REJECTED

Decision:
Do not introduce microservices for the initial product.

Reason:
A modular monolith is sufficient for the initial
scale and avoids unnecessary operational complexity.

## ADR-016 — User Interface Reference

Status: ACCEPTED

Decision:
The supplied ChaiCode Year screenshot is a
visual and interaction reference.

Consequences:
- Recreate behavior with original implementation.
- Do not copy proprietary source code.
- Preserve the user's subsequently supplied UI.
- Use the screenshot to validate layout hierarchy.

---

## New Decision Template

### ADR-XXX — Decision Title

Status:
PROPOSED / ACCEPTED / REJECTED / SUPERSEDED

Date:

Context:

Problem:

Options considered:

Decision:

Reason:

Consequences:

Affected files:

Related task IDs:

Migration impact:

Approved by:
