---
name: frontend-page-governor
description: Govern and incrementally refactor existing H5 pages in this repository. Use when the task is to analyze, clean up, or improve old `uni-app + Vue 3 + TypeScript` pages under `web/src/pages/*` without rewriting them, especially when files are too large, styles drift, states are inconsistent, or page sections should be normalized against `web/src/components/*` and `web/src/uni.scss`.
---

# Frontend Page Governor

Use this skill when the user wants to improve an existing H5 page in this repo without replacing the page wholesale.

This is a governance skill, not a redesign skill. Its job is to make old pages in `web/src/pages/*` easier to maintain, more consistent, and safer to extend.

## Stack And Scope

- Tech stack is fixed: `uni-app + Vue 3 + TypeScript`
- Page entry points live in `web/src/pages/*`
- Shared UI lives in `web/src/components/*`
- Shared style tokens/patterns live in `web/src/uni.scss`
- Common interaction feedback currently relies on `uni.showToast`, button `:loading`, conditional `v-if`, and page-local `ref` state

Treat this skill as project-private. Do not generalize it for React, Next, or other Vue projects.

## Mission

Given an existing page or page group:

1. Identify governance problems before touching code
2. Produce the smallest viable refactor path
3. Normalize structure, shared patterns, and page states
4. Preserve business logic and page behavior unless the user asked for behavior changes

## Hard Rules

- Do not rewrite the page from scratch
- Do not introduce new libraries
- Do not replace existing business flow just because the file is messy
- Do not create a new shared component until you have checked `web/src/components/*` and adjacent pages first
- Do not add one-off styles if the same pattern should live in `web/src/uni.scss` or an existing shared block
- Do not move logic across files unless the new boundary is clearer and reduces future duplication

## First Pass Workflow

1. Read the target page in `web/src/pages/*`
2. Check adjacent pages in the same module for repeated layout or state patterns
3. Check `web/src/components/*` for reusable building blocks before proposing extraction
4. Check `web/src/uni.scss` for existing spacing, colors, button, card, or form patterns that should be reused
5. Separate findings into structure, style consistency, and UX state issues

## What To Diagnose

### Structural Governance Issues

- page file is too large to reason about safely
- template, state, and handlers are tightly interleaved
- repeated sections could become local or shared components
- page contains duplicated formatting or conversion logic that should be centralized
- unrelated concerns are mixed in the same file

### UI Consistency Issues

- inconsistent spacing, card padding, section gaps, or form row rhythm
- repeated ad hoc colors, borders, shadows, or radii
- inconsistent title hierarchy or action placement
- same interaction pattern rendered differently across sibling pages

### State And UX Issues

- missing or weak `loading / empty / error / submitting` states
- duplicated state markup with slightly different wording or styling
- no disabled states while submitting
- success/failure feedback is inconsistent or missing
- error handling is silent or only logs without user feedback

## Required Output

When asked to analyze a page, produce these sections in order:

1. Problem list
2. Governance plan
3. Reuse opportunities
4. State normalization plan
5. Risk notes

Keep the plan incremental. Prefer phases such as:

- phase 1: split obvious sections without behavior change
- phase 2: normalize shared page states
- phase 3: lift repeated styles or blocks into shared patterns

## Governance Heuristics For This Repo

- If a block repeats only inside one page, prefer a local child component or an internal section split before promoting it to `web/src/components/*`
- If the same block appears across modules like `customer`, `product`, `boarding`, or `feeding`, check whether a shared component is justified
- If the inconsistency is mostly spacing, borders, and typography, prefer normalizing class structure and `web/src/uni.scss` before inventing a component
- If the page already follows a module-specific visual pattern, preserve it and only reduce drift
- If a page has primitive state handling but works, normalize it instead of replacing it with a new architectural pattern

## Escalate To Other Skills

- Use `frontend-component-extractor` when the main problem is file size or repeated sections
- Use `frontend-state-ux-normalizer` when the main problem is missing or inconsistent states and feedback
- Use `frontend-governed-implementer` only after governance direction is clear
