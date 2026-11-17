
# SCOPE.md

# Product Scope and Requirements

## 1. Scope Management

This document defines the approved product capabilities.

All implementation tasks must map to a requirement
in this document or an approved decision.

Do not introduce unrelated functionality.

## 2. MVP Modules

### M01 — Authentication

Requirements:
- User registration.
- User login.
- User logout.
- Session validation.
- Password hashing.
- Password recovery.
- Account profile.
- Session expiration.
- Session revocation.

Acceptance:
- Users can register and authenticate.
- Passwords are never stored in plaintext.
- Sessions are secure.
- Unauthorized API requests are rejected.

### M02 — User Management

Requirements:
- Profile management.
- Display name.
- Email.
- Avatar URL.
- Account status.
- Account preferences.
- Default timezone.

Acceptance:
- Users can update permitted profile fields.
- Users cannot modify their roles through profile APIs.

### M03 — Workspace Management

Requirements:
- Create workspace.
- Rename workspace.
- Workspace settings.
- Workspace owner.
- Workspace membership.
- Workspace switching.
- Workspace deletion.
- Ownership transfer.

Acceptance:
- Only authorized users manage workspaces.
- Ownership transfer requires explicit authorization.
- Membership removal revokes access.

### M04 — Invitations

Requirements:
- Invite by email.
- Select permitted role.
- Generate secure invitation token.
- Store token hash.
- Expiration.
- Revoke invitation.
- Accept invitation.
- Reject invitation.
- Prevent invitation replay.

Acceptance:
- Expired invitations cannot be accepted.
- Raw invitation tokens are never stored.
- Invitations cannot grant unauthorized privileges.

### M05 — RBAC

Requirements:
- Owner role.
- Admin role.
- Editor role.
- Member role.
- Viewer role.
- Custom roles.
- Permission catalogue.
- Permission overrides.
- Member role updates.

Acceptance:
- Every protected API checks permissions.
- Viewer cannot modify planner data.
- Editor cannot manage membership by default.
- Admin cannot perform owner-only actions by default.
- Permission changes take effect immediately.

### M06 — Annual Plans

Requirements:
- Create annual plan.
- Read annual plan.
- Update annual plan.
- Archive annual plan.
- Delete annual plan according to retention policy.
- Personal scope.
- Workspace scope.
- Year navigation.
- Plan timezone.

Acceptance:
- Each plan has a valid owner/scope.
- Users cannot access unauthorized plans.
- Duplicate plans follow the approved uniqueness policy.

### M07 — Calendar Engine

Requirements:
- ISO week calculations.
- Monday–Sunday week display.
- 52-week support.
- 53-week support.
- Leap-year support.
- Year crossover support.
- Timezone-aware dates.

Acceptance:
- Week identity is correct.
- December/January transitions are correct.
- Week 53 is handled when applicable.
- No date shifts due to browser timezone parsing.

### M08 — Year Board

Requirements:
- Annual overview.
- Quarter grouping.
- Month labels.
- Week cards.
- Seven-day indicators.
- Current week highlight.
- Year progress.
- Category filters.
- Activity visualization.

Acceptance:
- All weeks intersecting the selected year are visible.
- Board updates after relevant mutations.
- Filters work correctly.
- UI remains responsive.

### M09 — Weekly Planner

Requirements:
- Weekly focus.
- Weekly notes.
- Week navigation.
- Daily plan summaries.
- Weekly activity summary.
- Selected-week drawer.

Acceptance:
- Week data persists.
- Previous/next week navigation works.
- Drawer state is accessible.
- Closing the drawer preserves board context.

### M10 — Daily Planner

Requirements:
- Create daily plan item.
- Update item.
- Delete item.
- Complete item.
- Category.
- Optional time.
- Optional description.
- Ordering.

Acceptance:
- Date ownership is correct.
- Completed state persists.
- Daily items appear in the correct week.

### M11 — Habit Tracking

Requirements:
- Create habit.
- Update habit.
- Archive habit.
- Weekly schedule.
- Daily completion.
- Weekly target.
- Streak.
- Completion history.

Acceptance:
- Duplicate completion records are prevented.
- Streak logic is tested.
- Habit schedule is respected.
- Archived habits remain available in historical records.

### M12 — Quarterly Goals

Requirements:
- Q1 goals.
- Q2 goals.
- Q3 goals.
- Q4 goals.
- Goal status.
- Priority.
- Due date.
- Completion date.
- Progress summary.

Acceptance:
- Goals belong to the correct plan and quarter.
- Goal changes update relevant summaries.

### M13 — Activities

Requirements:
- Activity title.
- Category.
- Start time.
- End time.
- Duration.
- Date.
- Optional goal relation.
- Optional habit relation.
- Activity history.

Acceptance:
- Invalid durations are rejected.
- Date and timezone handling are consistent.
- Activity summaries are correct.

### M14 — Analytics

Requirements:
- Habit completion.
- Weekly progress.
- Goal completion.
- Activity totals.
- Quarter progress.
- Annual summary.

Acceptance:
- Analytics are calculated from authoritative data.
- Definitions of each metric are documented.
- Aggregations are tested.

### M15 — Audit

Requirements:
- Workspace member changes.
- Invitations.
- Role changes.
- Permission changes.
- Ownership transfer.
- Sensitive administrative operations.

Acceptance:
- Audit records identify actor and target.
- Audit logs cannot be edited by ordinary members.
- Sensitive tokens and credentials are excluded.

## 3. Future Scope

Do not implement these during the MVP without approval.

- Calendar provider integrations.
- Google Calendar synchronization.
- Microsoft Calendar synchronization.
- Recurring reminders.
- Email notifications.
- Push notifications.
- Mobile native applications.
- Offline-first synchronization.
- AI planning assistant.
- AI goal recommendations.
- Billing and subscriptions.
- Enterprise SSO.
- Advanced organization management.
- Public planner sharing.
- Advanced reporting exports.

## 4. Explicit Exclusions

Unless approved, do not implement:

- Microservices.
- Kubernetes.
- Event sourcing.
- CQRS infrastructure.
- GraphQL.
- Multiple databases.
- Redis.
- Kafka.
- Complex workflow engines.

These technologies must solve a demonstrated problem
before being introduced.

## 5. Functional Requirements

Every feature must support:

- Authentication.
- Input validation.
- Authorization.
- Correct persistence.
- Error handling.
- Loading states.
- Empty states.
- Success feedback.
- Tests.

## 6. Non-Functional Requirements

### Performance

The annual board should use one optimized summary
request rather than one request per week.

### Security

Tenant isolation must be enforced on the backend.

### Reliability

Database operations must be safe against retries
and duplicate requests where appropriate.

### Accessibility

Keyboard access and visible focus states are required.

### Maintainability

Keep domain logic separate from HTTP handlers.

### Compatibility

API contracts must remain compatible with
the supplied React frontend.

## 7. MVP Release Criteria

The MVP cannot be released until:

- Authentication works.
- Personal planner works.
- Year board works.
- Week details work.
- Habits work.
- Goals work.
- Workspace creation works.
- RBAC works.
- Tenant isolation tests pass.
- Production build passes.
- Backup and recovery procedures exist.
