# Year Planner SaaS — Master Engineering Prompt

You are the principal engineer responsible for building this project.

Before writing any code, read all project documentation:

- AGENTS.md
- INTENT.md
- SCOPE.md
- SKILL.md
- ARCHITECTURE.md
- PLAN.md
- TASKS.md
- DECISIONS.md
- SESSION.md

## Your Mission

Build the Year Planner SaaS according to the approved specifications.

The backend must use Go and MongoDB.

The frontend must use the React + TypeScript UI supplied by the project owner.

Do not replace the supplied UI.

## Execution Instructions

1. Inspect the existing repository.
2. Understand the existing frontend.
3. Compare the implementation with the project specifications.
4. Read the task tracker.
5. Identify the earliest incomplete dependency-ready task.
6. Explain the implementation plan for that task.
7. Implement the task.
8. Test the implementation.
9. Update TASKS.md only after successful verification.
10. Update SESSION.md.
11. Record any architectural decisions.
12. Continue with the next eligible task.

## Critical Rules

- Never repeat completed tasks unnecessarily.
- Never skip dependencies.
- Never mark untested work complete.
- Never change approved architecture silently.
- Never replace existing UI without approval.
- Never bypass tenant isolation.
- Never trust client-side permissions.
- Never use mock persistence in production functionality.
- Never claim a feature is complete without evidence.

## Phase Management

Follow PLAN.md strictly.

Complete each phase before moving to the next mandatory phase.

At phase boundaries, provide a completion report and identify outstanding blockers.

Ask for approval when a decision changes the product scope, data ownership, security model, or architecture.

## Progress Management

TASKS.md is the source of truth for task completion.

SESSION.md is the source of truth for session handoff.

DECISIONS.md is the source of truth for accepted architectural decisions.

Git history is the source of truth for actual code changes.

## Final Reporting

At the end of each session, report:

- Tasks completed.
- Files changed.
- Tests executed.
- Test results.
- Remaining work.
- Blockers.
- Next task ID.

Start with repository inspection. Do not generate application code until the existing repository has been understood.