---
name: feeding-module
description: Diagnose, implement, or review the 上门喂养 module in this repository. Use when the request involves feeding plans, visit execution, feeding pricing, playtime/陪玩, feeding settlement, feeding dashboard/calendar UI, or the feeding pages and APIs under `web/src/pages/feeding*`, `web/src/api/feeding.ts`, `server/internal/handler/feeding.go`, `server/internal/service/feeding_service.go`, and related feeding models/repositories.
---

# Feeding Module

Use this skill when the task touches the repo's 上门喂养 flow. This module is not a thin CRUD feature. It has its own plan lifecycle, visit lifecycle, pricing snapshot, settlement path, and mobile-heavy UI.

## Workflow

1. Identify which layer the request belongs to.
   - Plan creation/editing: `web/src/pages/feeding/create.vue`, `server/internal/service/feeding_service.go`
   - Plan detail / pause / resume / cancel / generate order: `web/src/pages/feeding/detail.vue`
   - Daily board or calendar/cards view: `web/src/pages/feeding/dashboard.vue`
   - Visit execution: `web/src/pages/feeding/visit-execute.vue`
   - Settings/pricing/items: `web/src/pages/feeding/settings.vue`
   - API surface: `web/src/api/feeding.ts`, `server/internal/handler/feeding.go`, `server/internal/router/router.go`

2. Confirm the data grain before changing UI or logic.
   - `FeedingPlan` is plan-level.
   - `FeedingVisit` is single-day execution-level.
   - If a request asks for day-level behavior, verify the data exists at visit level first. Do not fake day-level persistence from plan-level fields.

3. Check lifecycle constraints before patching behavior.
   - Plan statuses and visit statuses are defined in `server/internal/model/feeding.go`.
   - Order generation finalizes the plan and marks unfinished visits done.
   - Plan update is blocked for cancelled/completed plans.
   - Visit completion requires started status, at least one media item, and all checklist items checked.

4. For pricing or settlement issues, read backend pricing logic first.
   - `buildPlanPricing` computes per-day base charges, playtime charges, and other fees.
   - `GenerateOrder` converts completed visits plus plan-level extras into order items.
   - Use the backend as the source of truth; frontend only renders snapshots.

5. For UI bugs, verify in browser/mobile layout.
   - Prefer Playwright or real DOM inspection for dashboard/detail/visit-execute pages.
   - This module is used on mobile H5, so narrow viewport behavior matters.

6. After changing `web/` or `server/`, deploy with the repo hook.
   - Required command:
   ```bash
   printf '{"tool_input":{"file_path":"/absolute/path/to/changed/file"}}' | /Users/genglsh/workstation/cat/cat/.codex/hooks/deploy.sh
   ```

## Rules That Matter

- Treat `selected_dates_json` as the real schedule source for the current implementation.
- Treat `rules` as configuration/display support, not the primary execution source for currently generated visits.
- Do not assume playtime has per-day persistence. Current persisted fields are `play_mode` and `play_count` on `FeedingPlan`.
- Do not bypass `normalizePlanInput` when changing request shape. That method enforces customer/pet matching, date normalization, item validation, and address requirements.
- Do not change order-generation semantics lightly. Feeding orders are not the same as salon appointments or normal product orders.
- Preserve shop scoping and role restrictions. Staff-level users should only see/operate visits assigned to them unless the code explicitly widens visibility.

## Verification Checklist

- Backend change:
  - Read the affected service method in `server/internal/service/feeding_service.go`
  - Confirm repository preload/query shape in `server/internal/repository/feeding_repo.go`
  - If status or totals changed, verify with API or database query

- Frontend change:
  - Verify the exact page in H5 mobile viewport
  - Check whether the same data is rendered in both dashboard calendar and cards views when applicable
  - Check whether detail page, execute page, and order generation page still agree on amounts and status

- Cross-layer change:
  - Verify request payload shape in `web/src/api/feeding.ts`
  - Verify handler binding in `server/internal/handler/feeding.go`
  - Verify route registration in `server/internal/router/router.go`

## Read This Reference

Load [feeding-module-map.md](references/feeding-module-map.md) when you need:
- entity and state-flow details
- key file map
- pricing/order notes
- common pitfalls specific to this repo
