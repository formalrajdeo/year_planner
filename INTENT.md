
# INTENT.md

# Product Intent — Year Planner SaaS

## 1. Product Vision

Build a powerful annual planning platform that allows
individuals and teams to visualize, organize, track,
and execute their goals throughout an entire year.

The primary experience is a complete year displayed
within one planning workspace.

Users should understand:

- Where they are in the year.
- What they planned.
- What they completed.
- Which habits they maintained.
- Which goals are progressing.
- What remains unfinished.
- What they should focus on next.

The product must make annual planning visual,
interactive, measurable, and collaborative.

## 2. Product Philosophy

The product is built around five principles.

### 2.1 Year First

The year is the primary planning surface.

Users should not need to navigate through twelve
separate monthly pages to understand their year.

The year board must make all weeks visible
within a single desktop planning view.

### 2.2 Execution Over Decoration

The product is not merely a calendar.

It must connect:
- Goals.
- Weekly focus.
- Daily execution.
- Habits.
- Activities.
- Progress.

Every visual indicator should represent meaningful
planning or execution data.

### 2.3 Personal Ownership

Every user should be able to maintain an independent
personal planner.

Personal data must remain private by default.

### 2.4 Collaboration

Users should be able to create workspaces,
invite other members, and collaborate.

Workspace owners must control access through roles
and granular permissions.

### 2.5 Long-Term Reliability

Annual plans contain valuable personal and business
information.

Data integrity, privacy, authorization, backups,
and predictable behavior are essential.

## 3. Target Users

### Individual Users

People who want to:
- Organize their entire year.
- Track daily habits.
- Plan learning activities.
- Set quarterly goals.
- Review their progress.

### Professionals

People who want to:
- Organize work.
- Track important projects.
- Plan milestones.
- Maintain weekly focus.
- Balance personal and professional objectives.

### Teams

Teams that want to:
- Create shared plans.
- Assign members.
- Coordinate goals.
- Track activities.
- Monitor progress.

### Workspace Administrators

Administrators who want to:
- Manage members.
- Invite users.
- Assign roles.
- Configure permissions.
- Review audit activity.

## 4. Core User Experience

The application has one primary experience:

THE YEAR BOARD.

The board should communicate the complete annual
planning state without requiring excessive navigation.

### Year Board

Display:
- Selected year.
- Current week.
- Total weeks.
- Days elapsed.
- Days remaining.
- Overall year progress.
- Quarter grouping.
- Month grouping.
- Week cards.
- Daily activity markers.
- Activity category colors.
- Current week highlight.

### Sidebar

Display:
- Current date.
- Today's habits.
- Streak information.
- Weekly completion counts.
- Perfect week summary.
- Quarterly goals.
- Activity filters.

### Week Detail

Selecting a week opens its detail panel.

Display:
- Week number.
- Date range.
- Weekly focus.
- Habit completion matrix.
- Learning/activity summary.
- Daily plans.
- Notes.

The user must be able to move between weeks
without losing the main year-board context.

## 5. Product Ownership Model

The platform supports two independent planning scopes.

### Personal Scope

Each user can create private annual plans.

Personal plans:
- Belong to one user.
- Are private by default.
- Are not automatically shared.
- Are inaccessible to other workspace members.

### Workspace Scope

A workspace can own shared annual plans.

Workspace members may receive different permissions.

Workspace resources are owned by the workspace,
not automatically by the user who created them.

### Both Models

The product should support:

1. Personal annual plans.
2. Shared workspace annual plans.
3. Individual member plans within a workspace context.

Workspace membership must not automatically
grant access to a member's personal planner.

## 6. Product Differentiation

The product should focus on:

- Full-year visibility.
- Fast interaction.
- Weekly execution.
- Habit accountability.
- Quarterly goal tracking.
- Visual progress.
- Flexible collaboration.
- Granular permissions.
- Reliable data history.

## 7. Product Quality Goals

The experience should be:

- Fast.
- Responsive.
- Accessible.
- Predictable.
- Secure.
- Easy to understand.
- Suitable for long-term use.

The backend should remain lightweight,
using Go and MongoDB.

The frontend should preserve the supplied
React UI and integrate through documented APIs.

## 8. Product Success Criteria

The product is successful when users can:

1. Create an account.
2. Create their personal year plan.
3. View the complete year.
4. Select and edit weeks.
5. Create daily plans.
6. Track habits.
7. Create quarterly goals.
8. Review progress.
9. Create a workspace.
10. Invite members.
11. Configure permissions.
12. Collaborate securely.
13. Return later without losing progress.

## 9. Non-Goals

The initial product is not intended to be:

- A full enterprise ERP.
- A replacement for every project management tool.
- A social network.
- A general-purpose document editor.
- A complex financial management platform.

Future integrations may be evaluated separately.
