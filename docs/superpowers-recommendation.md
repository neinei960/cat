# Superpowers Skills Recommendations

This document records which `superpowers` skills are the best fit for this repository, when to use them, and how they should work with the repo's own `.codex/skills`.

## Goals

- Keep the workflow practical for this repo
- Prefer repo-specific skills for domain behavior
- Use `superpowers` to improve debugging, verification, planning, and review quality
- Avoid blindly applying heavyweight skills to small tasks

## Repo Context

This repository is a business system with coupled frontend and backend behavior:

- `web/` is the H5 frontend
- `server/` is the Go backend
- Many requests cross page UI, API payloads, service logic, and settlement rules
- The repo already requires deployment after `web/` or `server/` changes
- UI-affecting changes require browser verification after deployment

Because of that, the most valuable `superpowers` skills here are the ones that improve:

- root-cause debugging
- implementation planning for cross-layer changes
- final verification discipline
- review quality before merge
- skill authoring for repo-specific workflows

## Priority Tiers

### Tier 1: Recommended Often

These are the highest-value `superpowers` skills for day-to-day work in this repo.

#### `systematic-debugging`

Use when:

- a frontend page behaves incorrectly
- a backend endpoint returns wrong data
- order totals, discounts, or status flows look wrong
- a bug appears across `web/` and `server/`

Why it fits this repo:

- Many bugs here are cross-layer bugs
- Quick UI-only or handler-only patches often miss the real cause
- It pushes root-cause analysis before code changes

Best pairings:

- repo skill like `product-module`, `feeding-module`, or `boarding-module`
- browser verification after UI fixes

#### `verification-before-completion`

Use when:

- claiming a bug is fixed
- claiming a feature is complete
- preparing to close out a task

Why it fits this repo:

- This repo has a real deployment requirement
- UI changes must be checked in browser after deployment
- Final claims should be backed by actual commands and checks

Minimum expectation in this repo:

- run the relevant verification command
- deploy if `web/` or `server/` changed
- run a browser check if UI was affected

#### `requesting-code-review`

Use when:

- finishing a non-trivial feature
- changing pricing, settlement, order creation, status flows, or permissions
- completing a refactor that could introduce hidden regressions

Why it fits this repo:

- Business logic errors here are easy to miss in self-review
- Separate review is especially valuable for order math and lifecycle transitions

#### `writing-skills`

Use when:

- creating a new repo skill
- improving an existing repo skill
- codifying repeated workflows into `.codex/skills`

Why it fits this repo:

- This repo benefits from domain-specific skills
- Product, order, pricing, and operational flows are repo-specific
- Skills help keep future work consistent

Current good candidates:

- `product-module`
- future `order-module`
- future `pricing-module`
- future `dashboard-module`

### Tier 2: Use for Larger Tasks

These are useful, but not necessary for every task.

#### `writing-plans`

Use when:

- the task spans multiple subsystems
- the change requires backend, frontend, and verification coordination
- the request is large enough to justify a written execution plan

Good examples:

- redesigning retail product ordering
- adding an inventory subsystem
- refactoring order total calculation
- restructuring category or pricing models

#### `test-driven-development`

Use when:

- implementing backend logic changes
- changing deterministic business rules
- adding service-layer behavior that is testable in isolation

Why it is only Tier 2:

- This repo has significant UI-heavy work in Uni/H5
- Not every task has a clean test-first path
- It is still valuable for backend calculations and helpers

Best fit areas:

- Go service methods
- settlement math
- lifecycle validation logic
- repository-level query helpers when easy to cover

#### `subagent-driven-development`

Use when:

- you already have a plan
- tasks can be split into mostly independent units
- the work benefits from implementation plus staged review

Good examples:

- one task for frontend page changes
- one task for backend API/service updates
- one task for verification or regression review

#### `dispatching-parallel-agents`

Use when:

- there are 2 or more independent bugs
- several unrelated failures can be investigated separately
- you want fast parallel investigation instead of serial digging

Good examples:

- one product page bug and one feeding page bug at the same time
- one backend settlement issue and one mobile layout issue that do not share code

### Tier 3: Situational

These are useful in the right situation, but not core to the daily workflow here.

#### `receiving-code-review`

Use when:

- external review feedback arrives
- a comment seems technically questionable
- you need to validate review feedback against codebase reality before applying it

#### `finishing-a-development-branch`

Use when:

- the work is done
- tests and checks are complete
- you want a structured merge / PR / cleanup finish flow

This is more useful if branch hygiene and PR workflow become a stronger team norm here.

#### `using-git-worktrees`

Use when:

- parallel isolated workspaces are actually needed
- several large tasks must proceed on separate branches

Why not default:

- most work in this repo is faster in the current workspace
- worktrees add overhead if the task is small

#### `executing-plans`

Use when:

- there is already a written plan
- the goal is to execute it in a disciplined way

This is less common than `writing-plans` for everyday work.

## Low-Priority or Use with Caution

### `brainstorming`

This skill is intentionally heavyweight and requires design approval before implementation.

It can be useful for:

- ambiguous product work
- feature ideation
- cross-module redesign

It is not a good default for:

- straightforward bug fixes
- small UI patches
- normal repo maintenance

Reason:

- this repo often benefits more from direct implementation after brief codebase inspection
- the mandatory up-front design step is heavier than needed for many day-to-day tasks

### `using-superpowers`

Treat this as reference guidance, not the controlling workflow.

Reason:

- this repo already has strong system instructions, AGENTS guidance, and repo-local skills
- user instructions and repo requirements must stay higher priority
- native repo skills should remain the first choice for domain-specific work

## How Superpowers Should Combine with Repo Skills

Use repo skills for domain knowledge. Use `superpowers` for execution quality.

Recommended pattern:

1. Choose the repo skill for the module being changed
2. Add the relevant `superpowers` skill for process rigor
3. Implement
4. Deploy if `web/` or `server/` changed
5. Verify in browser if UI changed
6. Request review if the change is high-risk

Examples:

- Product bug:
  - `product-module` + `systematic-debugging`

- Product feature across frontend and backend:
  - `product-module` + `writing-plans`

- Settlement or commission change:
  - module skill + `test-driven-development` + `verification-before-completion`

- New repo skill creation:
  - `writing-skills`

## Recommended Default Workflow

For this repository, the default practical workflow is:

1. Use the repo-specific module skill first
2. If debugging, use `systematic-debugging`
3. If the task is large, use `writing-plans`
4. Before claiming completion, use `verification-before-completion`
5. For risky changes, use `requesting-code-review`
6. When codifying repeated knowledge, use `writing-skills`

## Suggested Task Mapping

### Product, order, pricing, feeding, boarding bugs

- repo module skill
- `systematic-debugging`
- `verification-before-completion`

### Medium to large feature work

- repo module skill
- `writing-plans`
- optionally `subagent-driven-development`
- `verification-before-completion`
- `requesting-code-review`

### Pure backend business logic changes

- repo module skill if applicable
- `test-driven-development`
- `verification-before-completion`

### Skill authoring and workflow codification

- `writing-skills`

## Summary

If only a few `superpowers` skills become part of the normal workflow for this repo, use these first:

- `systematic-debugging`
- `verification-before-completion`
- `requesting-code-review`
- `writing-plans`
- `writing-skills`

These provide the best return without fighting the repo's existing delivery style.
