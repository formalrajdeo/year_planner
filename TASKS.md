
# TASKS.md

# Project Execution Tracker

## 1. Tracking Rules

This file is the authoritative implementation ledger.

Every task must have:
- Stable task ID.
- Description.
- Dependencies.
- Acceptance criteria.
- Verification evidence.
- Status.

Statuses:

[ ] Not started
[~] In progress
[x] Completed and verified
[!] Blocked
[-] Deferred

Never mark a task complete before verification.

Never delete completed tasks.

Never repeat completed work without documenting why.

## 2. Current Project Status

Project Status: IN PROGRESS

Current Phase: Phase 2

Current Task: P2-009

Last Verified Task: P2-008

Last Updated: 2026-11-17

## 3. Phase 0 — Foundation

### Repository Discovery

- [x] P0-001 Inspect repository structure
- [x] P0-002 Inspect supplied React UI (none supplied; owner approved initial UI)
- [x] P0-003 Identify existing frontend routes (none present)
- [x] P0-004 Identify existing components (none present)
- [x] P0-005 Identify current dependencies (no frontend manifest present)
- [x] P0-006 Identify existing API integrations (none present)
- [x] P0-007 Document UI-to-feature mapping

### Backend Foundation

- [x] P0-008 Initialize Go module
- [x] P0-009 Create backend directory structure
- [x] P0-010 Configure environment validation
- [x] P0-011 Configure MongoDB client
- [x] P0-012 Configure connection pooling
- [x] P0-013 Implement HTTP server
- [x] P0-014 Configure graceful shutdown
- [x] P0-015 Configure request ID middleware
- [x] P0-016 Configure structured logging
- [x] P0-017 Implement liveness endpoint
- [x] P0-018 Implement readiness endpoint
- [x] P0-019 Configure centralized errors
- [x] P0-020 Configure OpenAPI

### Frontend Foundation

- [x] P0-021 Inspect frontend build configuration
- [x] P0-022 Create typed API client
- [x] P0-023 Configure TanStack Query
- [x] P0-024 Configure environment variables
- [x] P0-025 Configure API error handling
- [x] P0-026 Configure authentication state integration

### Quality Foundation

- [x] P0-027 Configure Go tests
- [x] P0-028 Configure frontend tests
- [x] P0-029 Configure linting
- [x] P0-030 Configure CI
- [x] P0-031 Configure Docker build
- [x] P0-032 Verify complete foundation — GitHub Actions CI passed after pushing commit 981b3af

## 4. Phase 1 — Authentication

- [x] P1-001 Create user domain model
- [x] P1-002 Create user repository
- [x] P1-003 Implement registration
- [x] P1-004 Implement password hashing
- [x] P1-005 Implement login
- [x] P1-006 Implement secure session creation
- [x] P1-007 Store session token hashes
- [x] P1-008 Implement session validation
- [x] P1-009 Implement logout
- [x] P1-010 Implement session revocation
- [x] P1-011 Implement authentication middleware
- [x] P1-012 Implement profile endpoint
- [x] P1-013 Implement profile update
- [x] P1-014 Implement password recovery (hashed one-time tokens, expiry, SMTP delivery via encrypted MongoDB outbox, session revocation, and UI flow)
- [x] P1-015 Implement MongoDB-backed login rate limiting (15-minute per-IP and per-IP/email limits; untrusted forwarding headers ignored)
- [x] P1-016 Implement CSRF protection (origin allowlist and session-bound double-submit token)
- [x] P1-017 Add authentication integration tests
- [x] P1-018 Verify authentication phase

Phase 1 verification: `go test ./...`, `go vet ./...`, `go test -race ./...`, frontend lint/typecheck/tests/build, and Docker image build passed. Tests cover one-time/expired reset tokens, session revocation, generic recovery responses, encrypted mail payloads, SMTP delivery, rate-limit thresholds, `Retry-After`, CSRF enforcement, and forged `X-Forwarded-For`. Local MongoDB startup/readiness passed with recovery disabled and with SMTP/outbox configured. The reset email provider still needs production environment values before deployment.

## 5. Phase 2 — Multi-Tenancy

- [x] P2-001 Create workspace model (typed settings object, owner, timestamps, soft-delete, schema version)
- [x] P2-002 Create workspace repository (global unique slugs)
- [x] P2-003 Create workspace API (authenticated routes and OpenAPI contract)
- [x] P2-004 Implement workspace creation
- [x] P2-005 Implement workspace listing
- [x] P2-006 Implement workspace update (owner-scoped rename)
- [x] P2-007 Implement membership model
- [x] P2-008 Implement membership repository
- [ ] P2-009 Implement membership listing
- [ ] P2-010 Implement invitation model
- [ ] P2-011 Implement secure invitation tokens
- [ ] P2-012 Implement invitation creation
- [ ] P2-013 Implement invitation acceptance
- [ ] P2-014 Implement invitation expiration
- [ ] P2-015 Implement invitation revocation
- [ ] P2-016 Implement membership removal
- [ ] P2-017 Implement workspace switching
- [ ] P2-018 Implement tenant scope helpers
- [ ] P2-019 Add personal data isolation tests
- [ ] P2-020 Add workspace isolation tests
- [ ] P2-021 Verify multi-tenancy phase (after P2-022 and P2-023)
- [ ] P2-022 Implement workspace soft deletion (after membership removal and tenant scope)
- [!] P2-023 Resolve and implement workspace ownership transfer

P2-001 verification: `go test ./...` and `go vet ./...` passed. Tests cover constructor defaults, required-field validation, and JSON/BSON field mappings.

P2-002 verification: `go test ./...`, `go vet ./...`, and the MongoDB-backed `TestMongoRepositoryTenantScopingAndGlobalSlugUniqueness` passed. The integration test verifies owner-scoped get/list, global slug conflict including soft-deleted workspaces, and index creation. API startup now creates the workspace indexes.

P2-003–P2-005 verification: `go test ./...` and `go vet ./...` passed. HTTP tests cover authenticated creation, trusted owner identity, CSRF, owner-scoped listing/retrieval, duplicate-slug conflicts, and storage failures. The OpenAPI YAML parsed successfully and documents the workspace endpoints. GitHub Actions CI run 44072914892 passed on commit `0d9816fa1194e32280cd105858ec8a4d89b351c1`.

P2-006 verification: `go test ./...` and `go vet ./...` passed. HTTP and MongoDB integration tests cover owner-only rename, not-found for cross-owner/deleted workspaces, unchanged slug and owner, updated timestamps, invalid names, and storage errors. OpenAPI YAML validation confirms the PATCH request accepts only `name`. GitHub Actions run 44074306056 passed on commit `ce6e86da059b1125afefbd5ae24eef08bb97476b`.

P2-007 completion evidence:
Implementation summary: Added the typed membership model, active/removed states, validation, and serialization coverage; documented the lifecycle without defining Phase 3 roles or permissions.
Files modified: `api/internal/memberships/membership.go`, `api/internal/memberships/membership_test.go`, `ARCHITECTURE.md`.
Tests executed: `go test ./...`; `go vet ./...`.
Test result: Passed.
Review status: Self-reviewed.
Completion date: 2026-11-17.

P2-008 completion evidence:
Implementation summary: Added validated membership creation and workspace/user-scoped lookup, a unique compound index, and startup index creation.
Files modified: `api/internal/memberships/repository.go`, `api/internal/memberships/mongo_repository.go`, `api/internal/memberships/mongo_repository_test.go`, `api/cmd/server/main.go`.
Tests executed: `go test ./...`; `go vet ./...`; MongoDB-backed `TestMongoRepositoryMembershipPersistenceAndUniqueness`.
Test result: Passed, including the opt-in integration test against the configured local MongoDB; GitHub Actions run 44078516910 passed on commit `c08d5aa2cee501be3435208cfdc1165da7f1e298`.
Review status: Self-reviewed; authorization checks are required by future service callers because lookups include removed records.
Completion date: 2026-11-17.

P2-022/P2-023 were added after reviewing approved product scope: workspace deletion and ownership transfer are required but had no implementation task.

P0-032 verification: GitHub Actions CI run 44070384027 passed on commit `981b3afc47712843dc679a70f299c1e4ac636561`, including Go checks, frontend checks, and Docker image build.

## 6. Phase 3 — RBAC

- [ ] P3-001 Define permission catalogue
- [ ] P3-002 Define Owner role
- [ ] P3-003 Define Admin role
- [ ] P3-004 Define Editor role
- [ ] P3-005 Define Member role
- [ ] P3-006 Define Viewer role
- [ ] P3-007 Implement permission resolver
- [ ] P3-008 Implement explicit grants
- [ ] P3-009 Implement explicit denials
- [ ] P3-010 Implement denial precedence
- [ ] P3-011 Implement delegation restrictions
- [ ] P3-012 Implement role assignment
- [ ] P3-013 Implement custom roles
- [ ] P3-014 Implement permission APIs
- [ ] P3-015 Implement audit events for role changes
- [ ] P3-016 Add authorization tests
- [ ] P3-017 Add privilege escalation tests
- [ ] P3-018 Verify RBAC phase

## 7. Phase 4 — Calendar and Annual Plans

- [ ] P4-001 Create annual plan model
- [ ] P4-002 Create annual plan repository
- [ ] P4-003 Implement personal plan creation
- [ ] P4-004 Implement workspace plan creation
- [ ] P4-005 Implement plan listing
- [ ] P4-006 Implement plan update
- [ ] P4-007 Implement plan archive
- [ ] P4-008 Implement timezone handling
- [ ] P4-009 Implement ISO week calculation
- [ ] P4-010 Implement ISO week-year
- [ ] P4-011 Implement 52-week year tests
- [ ] P4-012 Implement 53-week year tests
- [ ] P4-013 Implement leap-year tests
- [ ] P4-014 Implement December/January crossover tests
- [ ] P4-015 Implement year-board DTO
- [ ] P4-016 Implement year-board repository query
- [ ] P4-017 Implement year-board endpoint
- [ ] P4-018 Add annual plan authorization tests
- [ ] P4-019 Verify calendar phase

## 8. Phase 5 — Year Board UI

- [ ] P5-001 Map supplied year-board components
- [ ] P5-002 Map API fields to UI components
- [ ] P5-003 Integrate year-board query
- [ ] P5-004 Implement year navigation
- [ ] P5-005 Implement quarter grouping
- [ ] P5-006 Implement month grouping
- [ ] P5-007 Implement week cards
- [ ] P5-008 Implement seven-day cells
- [ ] P5-009 Implement current-week highlight
- [ ] P5-010 Implement activity indicators
- [ ] P5-011 Implement category filters
- [ ] P5-012 Integrate sidebar summaries
- [ ] P5-013 Implement loading states
- [ ] P5-014 Implement error states
- [ ] P5-015 Implement empty states
- [ ] P5-016 Verify responsive layout
- [ ] P5-017 Verify UI preservation

## 9. Phase 6 — Weekly and Daily Planner

- [ ] P6-001 Create week model
- [ ] P6-002 Implement week detail API
- [ ] P6-003 Implement week focus update
- [ ] P6-004 Implement week notes update
- [ ] P6-005 Create daily plan model
- [ ] P6-006 Implement daily plan repository
- [ ] P6-007 Implement daily plan creation
- [ ] P6-008 Implement daily plan update
- [ ] P6-009 Implement daily plan deletion
- [ ] P6-010 Implement completion state
- [ ] P6-011 Integrate week drawer
- [ ] P6-012 Implement previous/next week navigation
- [ ] P6-013 Implement autosave
- [ ] P6-014 Implement conflict handling
- [ ] P6-015 Verify weekly and daily phase

## 10. Phase 7 — Habits

- [ ] P7-001 Create habit model
- [ ] P7-002 Implement habit repository
- [ ] P7-003 Implement habit creation
- [ ] P7-004 Implement habit update
- [ ] P7-005 Implement habit archive
- [ ] P7-006 Implement habit scheduling
- [ ] P7-007 Create habit entry model
- [ ] P7-008 Implement unique daily entry constraint
- [ ] P7-009 Implement habit completion API
- [ ] P7-010 Implement weekly target calculation
- [ ] P7-011 Implement streak calculation
- [ ] P7-012 Implement perfect-week calculation
- [ ] P7-013 Implement habit history
- [ ] P7-014 Integrate habit matrix
- [ ] P7-015 Implement optimistic updates
- [ ] P7-016 Add habit calculation tests
- [ ] P7-017 Verify habit phase

## 11. Phase 8 — Goals and Activities

### Goals

- [ ] P8-001 Create goal model
- [ ] P8-002 Implement goal repository
- [ ] P8-003 Implement goal creation
- [ ] P8-004 Implement goal update
- [ ] P8-005 Implement goal deletion
- [ ] P8-006 Implement quarter assignment
- [ ] P8-007 Implement goal status
- [ ] P8-008 Implement goal completion

### Activities

- [ ] P8-009 Create activity model
- [ ] P8-010 Implement activity repository
- [ ] P8-011 Implement activity creation
- [ ] P8-012 Implement activity update
- [ ] P8-013 Implement activity deletion
- [ ] P8-014 Implement category management
- [ ] P8-015 Implement duration calculation
- [ ] P8-016 Implement goal/activity relations
- [ ] P8-017 Integrate activity bars
- [ ] P8-018 Integrate quarterly goals
- [ ] P8-019 Add aggregation tests
- [ ] P8-020 Verify goals and activities phase

## 12. Phase 9 — Analytics and Audit

- [ ] P9-001 Define analytics metrics
- [ ] P9-002 Implement weekly summaries
- [ ] P9-003 Implement quarterly summaries
- [ ] P9-004 Implement annual summaries
- [ ] P9-005 Implement habit analytics
- [ ] P9-006 Implement goal analytics
- [ ] P9-007 Implement activity analytics
- [ ] P9-008 Implement audit event model
- [ ] P9-009 Implement audit event repository
- [ ] P9-010 Implement administrative audit API
- [ ] P9-011 Add analytics tests
- [ ] P9-012 Verify audit phase

## 13. Phase 10 — Production

- [ ] P10-001 Review authentication security
- [ ] P10-002 Review authorization security
- [ ] P10-003 Test IDOR scenarios
- [ ] P10-004 Test cross-tenant access
- [ ] P10-005 Test privilege escalation
- [ ] P10-006 Review MongoDB indexes
- [ ] P10-007 Review slow queries
- [ ] P10-008 Perform API load testing
- [ ] P10-009 Perform frontend performance review
- [ ] P10-010 Configure monitoring
- [ ] P10-011 Configure error reporting
- [ ] P10-012 Configure database backups
- [ ] P10-013 Perform backup restore test
- [ ] P10-014 Review Docker image
- [ ] P10-015 Configure production secrets
- [ ] P10-016 Configure CI/CD
- [ ] P10-017 Document deployment
- [ ] P10-018 Document rollback
- [ ] P10-019 Perform accessibility review
- [ ] P10-020 Verify production release

## 14. Active Work Session

Current Phase:
Phase 2

Current Task:
P2-009

Assigned Agent:
Primary coding agent

Started:
Not started

Expected Outcome:
Workspace-scoped membership listing with appropriate authorization and API documentation.

## 15. Completion Evidence

For every completed task, record:

Task ID:
Implementation summary:
Files modified:
Tests executed:
Test result:
Review status:
Completion date:

A checked box without evidence is not accepted.

## 16. Blockers

Task ID: P2-023
Blocker: SCOPE.md requires workspace ownership transfer, but ARCHITECTURE.md does not define its endpoint or transfer/authorization policy.
Impact: Implementing transfer now would require guessing ownership, membership, and audit behavior.
Required decision: Approve the transfer API and state-transition rules.
Proposed resolution: Define a protected owner-transfer contract, validate the target member, and record the transfer in audit events before implementation.

When a task is blocked, record:

Task ID:
Blocker:
Impact:
Required decision:
Proposed resolution:

## 17. Change History

Maintain a short dated history of phase changes,
reopened tasks, deferred work, and major milestones.

Do not duplicate every code change here.
Use Git history for detailed source changes.
