---
name: frontend-component-extractor
description: Extract and normalize reusable sections from old H5 pages in this repository. Use when a `web/src/pages/*` file in this `uni-app + Vue 3 + TypeScript` project has grown too large, mixes unrelated sections, or repeats UI blocks that should become local child components or shared components under `web/src/components/*`, without changing business logic.
---

# Frontend Component Extractor

Use this skill when an old page is too large or structurally noisy, and the safest improvement is to split it into smaller sections.

This skill exists to improve maintainability. It is not permission to redesign the UI or rewrite the data flow.

## Scope

- Source pages: `web/src/pages/*`
- Shared components: `web/src/components/*`
- Shared styles: `web/src/uni.scss`
- Existing composition style: Vue SFCs with `script setup`, page-local `ref` state, direct API calls, and `uni` interactions

## Mission

1. Break a large page into understandable pieces
2. Decide the right extraction level for each piece
3. Keep behavior unchanged
4. Reduce duplication without over-abstracting

## Hard Rules

- Do not change business logic unless extraction forces a tiny signature cleanup
- Do not redesign page layout
- Do not create a "common" component for a block used once
- Do not move page-specific business rules into shared components
- Do not create prop-heavy abstractions just to satisfy DRY

## Extraction Levels

Pick the smallest level that improves the page:

1. Inline section split inside the same file
   - use when the page is messy but extraction would add file churn with little reuse

2. Module-local child component beside the page
   - use when a section is independent but only meaningful for that page or module

3. Shared component in `web/src/components/*`
   - use only when the same block or interaction clearly repeats across modules or is about to

## Workflow

1. Map the page sections
2. Mark repeated visual blocks, repeated form groups, repeated cards, repeated toolbars, repeated list rows, and repeated status blocks
3. For each candidate, decide:
   - is it page-local?
   - module-local?
   - truly shared?
4. Define the smallest prop/event interface possible
5. Keep API fetching, route handling, and business decisions in the page unless there is a clear existing pattern for lifting them out

## What To Extract In This Repo

Good extraction candidates:

- section headers with repeated action slots
- repeated filter bars or summary cards
- repeated list item cards
- repeated form sections with stable inputs
- repeated modal content blocks
- repeated state views such as loading/empty/error wrappers

Bad extraction candidates:

- a one-off layout block unique to one screen
- business-specific submit orchestration
- mixed blocks that still depend on half the page state
- tiny fragments that would increase file hopping without reducing complexity

## Required Output

When asked to split a page, return:

1. Component tree
2. File split plan
3. Responsibility of each file
4. Migration order
5. Over-abstraction risks

## Repo-Specific Guidance

- Prefer colocated module components when the block belongs to one business module such as `feeding`, `boarding`, `customer`, or `product`
- Only use `web/src/components/*` for genuinely shared H5 patterns
- If extraction would also require normalizing state or feedback, coordinate with `frontend-state-ux-normalizer`
- If multiple pages repeat the same spacing or wrapper pattern, consider style normalization before or alongside extraction
