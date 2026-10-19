---
name: frontend-state-ux-normalizer
description: Normalize page states and interaction feedback for existing H5 screens in this repository. Use when old `uni-app + Vue 3 + TypeScript` pages under `web/src/pages/*` have inconsistent or missing `loading`, `empty`, `error`, `submitting`, disabled, or toast feedback, and the goal is to improve UX with minimal layout change while aligning with existing `uni.showToast`, shared components, and `web/src/uni.scss`.
---

# Frontend State UX Normalizer

Use this skill when the page mostly works but feels fragile, silent, or inconsistent because state and feedback are incomplete.

This skill should improve UX without materially redesigning the page.

## Mission

1. Find missing page states
2. Normalize how those states are shown
3. Improve interaction feedback
4. Keep changes small and compatible with the repo's current patterns

## Hard Rules

- Do not introduce a new notification library
- Do not replace `uni.showToast` with a new global pattern unless the repo already has one
- Do not add complex skeleton systems unless the page already uses them
- Do not significantly move layout just to show state
- Do not hide backend failures; surface them with concise user-visible feedback

## State Model To Check

For each page or section, inspect whether these states exist and are distinct:

- initial loading
- refreshing or reloading
- empty result
- recoverable error
- submitting / saving / deleting
- success feedback
- disabled interactions during in-flight actions

## Interaction Model To Check

- primary buttons reflect in-flight state with `:loading` or disabled styles
- destructive actions have clear feedback
- list pages do not show a blank screen during fetch with no explanation
- detail pages do not silently fail when one request errors
- modal or inline actions surface success and failure with consistent wording

## Output Format

When asked to improve UX, provide:

1. Missing state inventory
2. Normalization plan
3. Code-level suggestions
4. Interaction checklist
5. Regression risks

## Repo-Specific Conventions

- Reuse existing page-local `loading`, `submitting`, and `list.length === 0` patterns where practical, but make naming and rendering more consistent
- Prefer concise `uni.showToast({ title, icon })` feedback for success and failure
- Prefer explicit empty/error blocks over implicit blank areas
- If many sibling pages repeat the same state markup, suggest a reusable state wrapper or shared section component
- Keep text tone aligned with current H5 wording in Chinese unless the user asked for copy changes

## Typical Fixes In This Repo

- normalize list page `loading / empty` views
- add explicit error banner or inline error block on detail pages
- disable repeat submit taps while saving
- add missing success/failure toasts for mutations
- make empty blocks consistent in spacing and typography

## Escalation

- If state normalization exposes large structural issues, involve `frontend-page-governor`
- If repeated state UI should become a component, involve `frontend-component-extractor`
- Use `frontend-governed-implementer` to land the agreed changes
