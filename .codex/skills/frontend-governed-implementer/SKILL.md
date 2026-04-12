---
name: frontend-governed-implementer
description: Implement incremental UI improvements in this repository after governance decisions are clear. Use when changing existing `uni-app + Vue 3 + TypeScript` H5 code under `web/src/pages/*`, `web/src/components/*`, or `web/src/uni.scss`, and the goal is to land minimal, maintainable refactors with normalized structure and UX instead of rewriting the page.
---

# Frontend Governed Implementer

Use this skill to apply the approved governance plan to the current repo's H5 frontend.

This skill is execution-only. It assumes the page has already been analyzed and the refactor direction is clear.

## Mission

1. Identify the smallest set of files to change
2. Reuse existing patterns before inventing new ones
3. Land minimal, maintainable refactors
4. Verify the real rendered result after deployment

## Hard Rules

- Do not rewrite a page from scratch
- Do not introduce new libraries
- Do not bypass existing shared components without a reason
- Do not duplicate styles that belong in `web/src/uni.scss` or an existing component
- Do not move backend or business logic unless the task explicitly requires it

## Required Read Path Before Editing

1. Read the target page in `web/src/pages/*`
2. Read adjacent pages in the same module for local consistency
3. Read candidate shared components in `web/src/components/*`
4. Read `web/src/uni.scss` if style normalization is in scope
5. Confirm whether the task is:
   - structure cleanup
   - component extraction
   - state normalization
   - style consistency
   - a mix of the above

## Implementation Priorities

Apply changes in this order:

1. reduce structural risk
2. remove duplication
3. normalize page states
4. normalize style usage
5. polish interaction feedback

Do not front-load polish while the page is still structurally unstable.

## File Placement Rules

- Shared cross-module UI goes in `web/src/components/*`
- Module-specific splits should stay near the page or within the same module structure
- Shared style tokens or utility classes belong in `web/src/uni.scss`
- Business-specific fetch/submit orchestration should stay in the page unless there is already a clear abstraction pattern

## Required Output

After implementation, report:

1. Files changed
2. What was normalized
3. What was intentionally left alone
4. Risks
5. Verification performed

## Verification Rules For This Repo

If you changed any file under `web/`, deployment is required.

Use:

```bash
printf '{"tool_input":{"file_path":"/absolute/path/to/changed/file"}}' | /Users/genglsh/workstation/cat/cat/.codex/hooks/deploy.sh
```

After deployment, run a Playwright check for UI-affecting changes. Verify the actual H5 rendering and the affected interaction flow, not just the static markup.

Minimum verification expectations:

- page loads without console/runtime errors
- loading/empty/error/submitting states render as intended
- primary actions still work
- extracted components receive the right data and events
- no obvious spacing or alignment regressions on H5 viewport

## When Not To Use This Skill

- If the user only asked for diagnosis, use `frontend-page-governor`
- If the main work is deciding extraction boundaries, use `frontend-component-extractor`
- If the main work is identifying missing states and feedback, use `frontend-state-ux-normalizer`
