
# PLAN.md

# Project Execution Roadmap

## 1. Execution Strategy

Development is organized into sequential phases.

Each phase contains:
- Objectives.
- Dependencies.
- Tasks.
- Deliverables.
- Acceptance criteria.
- Testing requirements.
- Exit criteria.

A phase is not complete until its exit criteria pass.

## 2. Phase 0 — Repository Discovery and Foundation

Objective:
Understand the existing UI and establish the project foundation.

Tasks:
- Inspect repository.
- Inspect supplied React UI.
- Identify existing components.
- Identify existing routes.
- Identify current dependencies.
- Document frontend structure.
- Create Go API module.
- Configure environment loading.
- Configure MongoDB client.
- Configure HTTP server.
- Configure request IDs.
- Configure logging.
- Configure health endpoints.
- Configure frontend API client.
- Configure OpenAPI.
- Configure tests.
- Configure CI.

Deliverables:
- Working React frontend.
- Running Go API.
- MongoDB connection.
- Environment validation.
- CI pipeline.
- Initial architecture documentation.

Exit criteria:
- Go server starts.
- Database readiness check works.
- Frontend build passes.
- API health check passes.
- Tests run in CI.

## 3. Phase 1 — Authentication and Users

Dependencies:
Phase 0.

Objective:
Establish trusted user identity.

Tasks:
- User model.
- User repository.
- Registration.
- Password hashing.
- Login.
- Session generation.
- Session hashing.
- Session validation.
- Logout.
- Session revocation.
- Profile endpoint.
- Password recovery.
- Authentication middleware.
- Authentication rate limiting.
- CSRF protection.

Deliverables:
Working authentication.

Exit criteria:
- Registration works.
- Login works.
- Logout revokes session.
- Password hashing verified.
- Unauthorized requests rejected.
- Authentication tests pass.

## 4. Phase 2 — Workspaces and Multi-Tenancy

Dependencies:
Phase 1.

Objective:
Implement personal and workspace ownership.

Tasks:
- Workspace collection.
- Membership collection.
- Workspace creation.
- Workspace listing.
- Workspace update.
- Membership creation.
- Member listing.
- Invitation model.
- Secure invitation tokens.
- Invitation acceptance.
- Invitation revocation.
- Membership removal.
- Workspace switching.
- Tenant scope abstraction.
- Tenant isolation tests.

Deliverables:
Working workspace system.

Exit criteria:
- Users can create workspaces.
- Users can join invited workspaces.
- Personal data remains private.
- Cross-workspace access is denied.
- Member removal revokes access.

## 5. Phase 3 — RBAC and Permissions

Dependencies:
Phase 2.

Objective:
Implement secure fine-grained authorization.

Tasks:
- Permission catalogue.
- Default role definitions.
- Permission resolver.
- Permission middleware/helpers.
- Role assignment.
- Member permission overrides.
- Custom role model.
- Delegation rules.
- Permission management APIs.
- Audit events for role changes.
- Authorization integration tests.

Deliverables:
Working RBAC.

Exit criteria:
- Viewer cannot write.
- Editor cannot manage members.
- Admin cannot perform owner-only operations.
- Custom permissions work.
- Privilege escalation tests pass.

## 6. Phase 4 — Calendar Engine and Annual Plans

Dependencies:
Phase 2 and Phase 3.

Objective:
Implement correct calendar calculations and plan ownership.

Tasks:
- Annual plan model.
- Personal plan creation.
- Workspace plan creation.
- Plan listing.
- Plan update.
- Plan archive.
- Timezone support.
- ISO week engine.
- ISO week-year calculation.
- Week 53 support.
- Leap-year tests.
- Cross-year tests.
- Year-board response DTO.
- Year-board repository queries.
- Year-board API.

Deliverables:
Complete year-board data API.

Exit criteria:
- Correct 52/53 week calculations.
- Correct ISO week identity.
- Tenant-safe plan APIs.
- Year-board response is compact.
- Calendar tests pass.

## 7. Phase 5 — Year Board UI Integration

Dependencies:
Phase 4.

Objective:
Integrate the Go API with the supplied React UI.

Tasks:
- Inspect supplied year-board components.
- Map components to API fields.
- Integrate API client.
- Integrate TanStack Query.
- Integrate year navigation.
- Integrate current-week highlight.
- Integrate quarter grouping.
- Integrate week cards.
- Integrate activity markers.
- Integrate filters.
- Integrate sidebar summaries.
- Add loading states.
- Add error states.
- Add empty states.
- Add responsive behavior.

Deliverables:
Functional annual board.

Exit criteria:
- Real database data appears.
- All weeks render.
- Year switching works.
- Current week is highlighted.
- Filters work.
- Existing UI visual design is preserved.

## 8. Phase 6 — Weekly and Daily Planner

Dependencies:
Phase 5.

Objective:
Enable weekly execution.

Tasks:
- Week collection.
- Week focus.
- Week notes.
- Week detail endpoint.
- Week update endpoint.
- Daily plan model.
- Daily plan CRUD.
- Daily item completion.
- Weekly drawer integration.
- Previous/next week navigation.
- Autosave.
- Conflict handling.
- Integration tests.

Deliverables:
Functional week drawer and daily planner.

Exit criteria:
- Focus persists.
- Notes persist.
- Daily items persist.
- Week navigation works.
- Board summaries refresh.

## 9. Phase 7 — Habit Tracking

Dependencies:
Phase 6.

Objective:
Implement habits and accountability.

Tasks:
- Habit CRUD.
- Habit scheduling.
- Weekly targets.
- Habit entries.
- Unique entry constraints.
- Daily completion.
- Weekly completion summary.
- Streak calculation.
- Perfect-week calculation.
- Archive behavior.
- Habit matrix UI.
- Optimistic updates.
- Tests for all streak rules.

Deliverables:
Functional habit tracker.

Exit criteria:
- Habit check-ins persist.
- Duplicate check-ins prevented.
- Streaks are correct.
- Weekly counts are correct.
- Historical entries remain available.

## 10. Phase 8 — Quarterly Goals and Activities

Dependencies:
Phase 6.

Objective:
Connect planning with measurable progress.

Tasks:
- Goal CRUD.
- Quarter assignment.
- Goal status.
- Priority.
- Due dates.
- Completion dates.
- Activity CRUD.
- Activity categories.
- Duration calculation.
- Goal/activity relationships.
- Quarterly summaries.
- Activity bars.
- Goal sidebar integration.
- Tests.

Deliverables:
Functional goals and activity tracking.

Exit criteria:
- Quarter grouping works.
- Goal progress is correct.
- Activity duration is correct.
- Annual summaries are correct.

## 11. Phase 9 — Analytics and Audit

Dependencies:
Phase 7 and Phase 8.

Objective:
Provide reliable progress reporting and accountability.

Tasks:
- Weekly analytics.
- Quarterly analytics.
- Annual analytics.
- Habit analytics.
- Activity analytics.
- Goal analytics.
- Audit event creation.
- Audit event listing.
- Administrative audit UI.
- Aggregation tests.

Deliverables:
Analytics and audit capabilities.

Exit criteria:
- Metrics have documented definitions.
- Aggregations are tested.
- Audit records identify actor and target.
- Sensitive information is excluded.

## 12. Phase 10 — Production Hardening

Dependencies:
All previous phases.

Objective:
Prepare the application for real users.

Tasks:
- Security review.
- IDOR testing.
- RBAC review.
- Load testing.
- Database index review.
- Slow query review.
- API latency review.
- Error monitoring.
- Backup configuration.
- Restore drill.
- Graceful shutdown verification.
- Docker image review.
- Secret management.
- CI/CD release workflow.
- Deployment documentation.
- Rollback documentation.
- Accessibility review.
- Cross-browser testing.

Deliverables:
Production release candidate.

Exit criteria:
- No unresolved critical security issues.
- Tenant isolation verified.
- Tests pass.
- Backups are restorable.
- Deployment and rollback are documented.
- Monitoring is operational.

## 13. Phase Advancement Rules

A phase may advance only when:

1. All mandatory tasks are completed.
2. Tests pass.
3. Acceptance criteria pass.
4. Documentation is updated.
5. Known blockers are resolved or explicitly accepted.
6. Security-sensitive work is reviewed.

## 14. Parallel Work Rules

Parallelize only independent tasks.

Examples:
- API documentation and independent UI inspection.
- Independent unit tests.
- Separate feature modules after shared contracts exist.

Never parallelize conflicting changes to:
- Permission engine.
- Shared domain models.
- MongoDB ownership model.
- Core API contracts.

## 15. Phase Handoff

At the end of each phase, create a report:

Phase:
Status:
Completed task IDs:
Modified files:
Tests:
Known limitations:
Decisions:
Next phase:
Outstanding blockers:

Update TASKS.md and DECISIONS.md as needed.
