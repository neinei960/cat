# Customer Detail Member Record Collapse Design

## Context

The customer detail page currently renders the member-card record list in an always-open state and shows up to three records by default. The requested change is to reduce visual noise on mobile while still keeping the most recent balance movement visible.

Primary surface:

- `web/src/pages/customer/detail.vue`

## Goal

Change the member-card record area so that:

- the latest record stays visible by default
- the rest of the records stay hidden until the user expands the section
- the expand/collapse control is a downward arrow style trigger instead of the current text link

## Non-Goals

- No backend or API changes
- No changes to record edit/delete business rules
- No changes to record ordering or record content
- No changes to non-member-card sections on the customer detail page

## Chosen Approach

Use a single local expand/collapse state on the customer detail page.

Behavior:

- When `records.length === 0`, keep the current behavior and render nothing for the record section.
- When `records.length === 1`, show the single record and omit the arrow toggle.
- When `records.length > 1`, show only the first record while collapsed.
- Clicking the arrow toggles between collapsed and expanded states.
- Expanded state shows the full `records` array in its current order.

## UI Design

Header:

- Keep the title `充值/消费记录`.
- Replace the current `查看全部/收起` text with a compact arrow trigger on the right.
- The trigger should visually communicate state:
  - collapsed: down arrow
  - expanded: up arrow
- A light count label may sit next to the arrow to indicate total record count without overpowering the header.

List:

- Reuse the existing record row layout and styles.
- Preserve admin edit/delete actions exactly as they work today.
- Do not introduce modal, accordion animation, or cross-section layout changes.

## State and Data Flow

- Reuse the existing `showAllRecords` boolean or rename it to a clearer local UI state if needed.
- Update `displayRecords` so it returns:
  - `records.slice(0, 1)` while collapsed
  - `records` while expanded
- Keep record fetch and refresh flow unchanged.

## Edge Cases

- If records are reloaded after edit/delete, the current expanded/collapsed state should remain stable.
- If deletion reduces the list from multiple records to one, hide the arrow automatically.
- If the latest record is edited, the visible first row should reflect refreshed data without extra UI handling.

## Verification Plan

Code-level:

- Add a small regression script that asserts the collapsed-state computed logic only exposes one record by default and all records when expanded.

UI-level:

- Deploy the changed H5 page with the repo deploy hook.
- Verify on the customer detail page that:
  - one record is visible by default when multiple records exist
  - arrow toggle expands to all records
  - arrow toggle collapses back to one record
  - edit/delete controls still appear for eligible records in expanded state

## Risks

- The current page already contains several local UI states; the change should stay isolated to the records section to avoid accidental regressions in member-card dialogs.
- If the arrow affordance is too subtle, users may miss that the section is expandable, so the count label should remain visible when there are multiple records.
