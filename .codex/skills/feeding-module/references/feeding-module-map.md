# Feeding Module Map

## Scope

The 上门喂养 module covers:
- feeding settings
- feeding plan creation/editing
- daily dashboard/calendar/cards
- per-visit execution
- feeding plan settlement into an order

It is separate from salon appointments. Do not reuse appointment assumptions here.

## Core Entities

Defined in `/Users/genglsh/workstation/cat/cat/server/internal/model/feeding.go`.

### `FeedingPlan`
- One customer-level plan across a date range
- Important fields:
  - `start_date`, `end_date`
  - `selected_dates_json`
  - `selected_items_json`
  - `play_mode`, `play_count`
  - `other_price`
  - `deposit`, `total_amount`, `unpaid_amount`
  - `status`
- Child records:
  - `Pets`
  - `Rules`
  - `Visits`

### `FeedingVisit`
- One scheduled execution unit
- Current implementation creates one visit per selected date, with `window_code=all_day`
- Important fields:
  - `scheduled_date`
  - `staff_id`
  - `status`
  - `visit_price`
  - `arrived_at`, `completed_at`
  - `customer_note`, `internal_note`, `exception_type`

### `FeedingVisitItem`
- Checklist items derived from selected feeding item templates
- Must all be checked before visit completion

### `FeedingVisitMedia`
- Execution evidence
- At least one image is required before completing a visit

## Current State Machines

### Plan statuses
- `draft`
- `active`
- `paused`
- `completed`
- `cancelled`

Main transitions:
- create -> `active`
- active -> paused
- paused -> active
- active/paused -> cancelled
- generate order -> `completed`

Important guardrails:
- cancelled/completed plans cannot be edited
- cancelling fails if any visit is `in_progress`

### Visit statuses
- `pending`
- `assigned`
- `in_progress`
- `done`
- `exception`
- `cancelled`

Main transitions:
- create visit -> `pending`
- assign -> `assigned`
- start -> `in_progress`
- complete -> `done`
- exception -> `exception`
- cancel plan -> pending/assigned visits become `cancelled`

Important guardrails:
- only pending/assigned can start
- only in_progress can complete
- only assigned/in_progress can be marked exception
- completing requires:
  - `arrived_at`
  - at least 1 media item
  - all checklist items checked

## Pricing and Settlement

Main logic: `/Users/genglsh/workstation/cat/cat/server/internal/service/feeding_service.go`

### Pricing build
Method: `buildPlanPricing`

Composition:
- per-day base price using holiday/non-holiday rules
- daily or counted playtime charge
- `other_price`

Stored snapshots:
- `pricing_snapshot_json`
- `selected_items_json`
- `selected_dates_json`

### Order generation
Method: `GenerateOrder`

Behavior:
- if plan already has an order, returns existing order
- unfinished visits are auto-marked `done` before settlement
- only `done` visits contribute to final order grouping
- base fee is grouped by daily price label
- playtime is added as plan-level order items, not inferred from visit items
- `other_price` becomes its own order item
- plan becomes `completed`

Practical implication:
- settlement bugs usually belong to backend service logic, not frontend totals

## Frontend Map

### Settings
- `/Users/genglsh/workstation/cat/cat/web/src/pages/feeding/settings.vue`
- pricing and default item template maintenance

### Create/Edit plan
- `/Users/genglsh/workstation/cat/cat/web/src/pages/feeding/create.vue`
- uses `selected_dates`
- sends `play_mode`, `play_count`, `other_price`
- validates through backend normalization

### Dashboard
- `/Users/genglsh/workstation/cat/cat/web/src/pages/feeding/dashboard.vue`
- two main views:
  - calendar
  - cards
- builds card rows from `FeedingPlan`, not `FeedingVisit`

### Detail
- `/Users/genglsh/workstation/cat/cat/web/src/pages/feeding/detail.vue`
- displays plan summary
- edits deposit
- pause/resume/cancel/generate order
- shows visits and notes

### Visit execute
- `/Users/genglsh/workstation/cat/cat/web/src/pages/feeding/visit-execute.vue`
- uploads media
- marks checklist items
- starts/completes/flags exception

### Shared helpers
- `/Users/genglsh/workstation/cat/cat/web/src/utils/feeding.ts`
- labels, date range expansion, snapshot parsing

## API Surface

Frontend wrapper:
- `/Users/genglsh/workstation/cat/cat/web/src/api/feeding.ts`

Backend handler/router:
- `/Users/genglsh/workstation/cat/cat/server/internal/handler/feeding.go`
- `/Users/genglsh/workstation/cat/cat/server/internal/router/router.go`

Key routes:
- `GET /b/feeding/settings`
- `PUT /b/feeding/settings/pricing`
- `PUT /b/feeding/settings/items`
- `POST /b/feeding/plans`
- `GET /b/feeding/plans`
- `GET /b/feeding/plans/:id`
- `PUT /b/feeding/plans/:id`
- `PUT /b/feeding/plans/:id/pause`
- `PUT /b/feeding/plans/:id/resume`
- `PUT /b/feeding/plans/:id/cancel`
- `POST /b/feeding/plans/:id/generate-order`
- `PUT /b/feeding/plans/:id/deposit`
- `GET /b/feeding/dashboard`
- `GET /b/feeding/visits`
- `PUT /b/feeding/visits/:id/assign`
- `PUT /b/feeding/visits/:id/note`
- `PUT /b/feeding/visits/:id/start`
- `PUT /b/feeding/visits/:id/complete`
- `PUT /b/feeding/visits/:id/exception`
- `POST /b/feeding/visits/:id/media`

## Repository Notes

Main repository:
- `/Users/genglsh/workstation/cat/cat/server/internal/repository/feeding_repo.go`

Important preload shape:
- plan detail preloads customer, order items, pets, rules, visits, visit staff, visit items, visit logs, visit media
- dashboard/listing often works from plan-level data, not just visits

## Known Pitfalls

1. Day-level vs plan-level confusion
- `play_mode` and `play_count` are plan-level fields.
- There is no persisted “which date has playtime” table in current schema.
- If UI asks for a per-date play marker, verify whether it is only presentational.

2. `selected_dates_json` is the operative schedule
- Current pricing and visit generation depend on selected dates.
- Do not assume weekday/window rules are the active source of visit generation.

3. Completion requirements are strict
- Missing media or unchecked items will block completion.
- If the UI says completion failed, inspect visit media and visit items first.

4. Order generation mutates visit state
- Generating an order can auto-finish unfinished visits.
- This can surprise UI assumptions if you only inspect frontend state.

5. Mobile H5 is the primary UI reality
- Some pages may work in desktop browser but fail in phone viewport.
- For dashboard/detail/visit execution, test in narrow width before declaring a UI fix done.

## Useful Greps

```bash
rg -n "feeding|Feeding|上门喂养" web server -g '!node_modules'
rg -n "play_mode|play_count|selected_dates_json|generate-order" web server -g '!node_modules'
rg -n "FeedingPlanStatus|FeedingVisitStatus" server/internal/model/feeding.go
```
