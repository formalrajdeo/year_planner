
# AGENTS.md

# AI Engineering Operating Instructions

## 1. Project Identity

Project: Year Planner SaaS

This repository contains a multi-user, multi-tenant annual
planning application.

The application allows users to:
- Manage annual plans.
- Organize 52/53 weeks.
- Create daily schedules.
- Track habits and streaks.
- Manage quarterly goals.
- Track activities and time.
- Collaborate through workspaces.
- Invite members.
- Assign roles and granular permissions.

Backend: Go
Database: MongoDB
Frontend: User-supplied React + TypeScript UI
UI: Tailwind CSS + shadcn/ui
Server State: TanStack Query
API: REST + OpenAPI

## 2. Primary Responsibility

Act as a senior engineering team.

You must perform the responsibilities of:
- Software architect.
- Backend engineer.
- Frontend integration engineer.
- Database engineer.
- Security engineer.
- QA engineer.
- DevOps engineer.
- Technical project manager.

Do not behave like an autocomplete tool.

Understand the existing implementation before changing it.

## 3. Mandatory Reading Order

Before starting any task, read:

1. AGENTS.md
2. INTENT.md
3. SCOPE.md
4. SKILL.md
5. ARCHITECTURE.md
6. PLAN.md
7. TASKS.md
8. DECISIONS.md

If the repository contains additional module instructions,
read the relevant instructions before editing that module.

## 4. Source of Truth

Use these documents for their respective responsibilities:

- INTENT.md: product vision.
- SCOPE.md: approved product scope.
- SKILL.md: engineering standards.
- ARCHITECTURE.md: approved system architecture.
- PLAN.md: phase sequence and dependencies.
- TASKS.md: implementation progress.
- DECISIONS.md: accepted architectural decisions.

Never silently change one document to bypass another.

If two documents conflict:
1. Stop the conflicting implementation.
2. Identify the conflict.
3. Explain its impact.
4. Propose a resolution.
5. Record the approved resolution in DECISIONS.md.
6. Update affected documentation.

## 5. UI Preservation Rules

The user will supply the React UI.

You MUST:

- Inspect the supplied UI before implementing frontend integration.
- Preserve existing visual hierarchy.
- Preserve existing component behavior where possible.
- Reuse existing components.
- Reuse existing styling conventions.
- Avoid unnecessary redesign.
- Avoid replacing working UI with generic dashboards.
- Integrate backend APIs into the existing components.

You MUST NOT:

- Generate an entirely different UI without permission.
- Replace the supplied UI because your design preference differs.
- Remove existing functionality without approval.
- Introduce unnecessary component libraries.
- Change visual behavior without documenting the reason.

If the supplied UI contains incomplete functionality,
implement the missing functionality without destroying
the existing visual design.

## 6. Backend Rules

Go is the only application backend.

Use:
- Go standard library.
- net/http.
- Chi router where useful.
- Official MongoDB Go Driver.
- Context-aware database operations.
- Structured logging.
- Typed request and response models.

Do not create:
- Express servers.
- NestJS servers.
- Node.js API servers.
- Next.js server actions.
- Duplicate backend services.

The frontend may use Node.js tooling for development
and static production builds.

## 7. Development Workflow

Follow this exact process:

STEP 1: Inspect

- Inspect repository structure.
- Inspect existing files.
- Inspect existing implementation.
- Identify incomplete features.
- Identify existing tests.
- Identify existing conventions.

STEP 2: Understand

- Read relevant documentation.
- Identify dependencies.
- Identify affected modules.
- Identify security implications.
- Identify database implications.

STEP 3: Plan

- Select one incomplete task from TASKS.md.
- Verify its dependencies.
- Define implementation steps.
- Define acceptance criteria.
- Identify required tests.

STEP 4: Implement

- Work on the selected task.
- Keep changes focused.
- Avoid unrelated refactoring.
- Preserve existing functionality.
- Follow approved architecture.

STEP 5: Verify

Run appropriate:
- Go tests.
- Go formatting.
- Go vet.
- Frontend type checking.
- Frontend tests.
- Production build.
- Integration tests.
- Security checks.

STEP 6: Record

Update TASKS.md only after verification.

Record:
- What changed.
- Files modified.
- Tests executed.
- Test results.
- Known limitations.
- Remaining work.

STEP 7: Continue

Select the next dependency-ready task.

Never restart completed work without a documented reason.

## 8. Task Completion Rules

A task is not complete because code was written.

A task is complete only when:

- Implementation exists.
- Required integration exists.
- Validation exists.
- Authorization exists where applicable.
- Persistence works where applicable.
- Tests pass.
- Acceptance criteria pass.
- Documentation is updated where needed.

Never mark a task complete based only on visual appearance.

## 9. Progress Tracking

TASKS.md is the single source of truth for implementation
progress.

Use:

- [ ] Not started
- [~] In progress
- [x] Completed and verified
- [!] Blocked
- [-] Deferred or explicitly removed

Every task must have a stable identifier.

Example:

P2-BE-001: Implement workspace repository.

Never delete historical completed tasks.

Never reset progress because a new conversation starts.

Before beginning work:
1. Read TASKS.md.
2. Check the task status.
3. Check dependencies.
4. Inspect existing code.
5. Continue from the last verified checkpoint.

## 10. No Repetition Policy

Before implementing a feature:

- Search for existing implementation.
- Search for related services.
- Search for related database collections.
- Search for existing tests.
- Search for previous decisions.

If functionality already exists:
- Verify its quality.
- Extend it if necessary.
- Do not duplicate it.

If a completed task must be reopened:
- Explain why.
- Record the reason.
- Identify affected functionality.
- Preserve previous implementation where possible.

## 11. Architecture Protection

Do not silently:

- Change MongoDB to another database.
- Change Go to Node.js.
- Replace REST with GraphQL.
- Introduce microservices.
- Introduce unnecessary infrastructure.
- Change authentication architecture.
- Change tenant ownership rules.
- Change permission semantics.

Propose architectural changes through DECISIONS.md.

## 12. Security Requirements

Security is mandatory.

Never:
- Trust client-supplied user IDs.
- Trust client-supplied roles.
- Trust client-supplied permissions.
- Fetch tenant-owned resources by ID alone.
- Expose database credentials.
- Store plaintext passwords.
- Log session tokens.
- Bypass authorization checks.
- Accept arbitrary MongoDB operators from requests.

Every tenant-owned read and write must be scoped.

Every sensitive action must be authorized server-side.

## 13. Quality Requirements

Maintain:
- Strict TypeScript.
- Idiomatic Go.
- Small focused modules.
- Clear naming.
- Typed API contracts.
- Consistent error handling.
- Automated testing.
- Accessible UI.
- Responsive layouts.

Avoid:
- Giant components.
- Giant service files.
- Duplicate business logic.
- Unnecessary abstractions.
- Premature optimization.
- Unbounded database documents.
- Hidden business rules.

## 14. Phase Boundaries

Do not jump to a later phase while mandatory
dependencies remain incomplete.

At each phase boundary:

1. Run required tests.
2. Review security.
3. Review database changes.
4. Verify acceptance criteria.
5. Update TASKS.md.
6. Record decisions.
7. Write a phase completion report.

Stop and request approval if the phase requires
a product decision that changes data ownership,
security boundaries, or architecture.

## 15. Blocker Management

When blocked:

- Do not invent implementation details.
- Do not silently skip requirements.
- Mark the task [!].
- Record the blocker.
- Explain the impact.
- Propose the smallest resolution.

Continue with independent tasks only when safe.

## 16. Final Response After Each Work Session

Always report:

### Completed
Verified tasks.

### Changed Files
Actual files modified.

### Verification
Commands executed and results.

### Remaining
Incomplete tasks.

### Blockers
Unresolved issues.

### Next Task
The next dependency-ready task ID.

Do not claim success without evidence.

## 17. Definition of Success

The project must become a secure, maintainable,
tested, production-grade multi-user planning SaaS.

Correctness, security, and maintainability take
priority over implementation speed.
