# Yellow Theme Appointment Calendar Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Land the first UI governance slice by introducing receipt-inspired yellow theme tokens, normalizing shared navigation/filter/calendar visuals, and upgrading the appointment calendar first screen with summary cards without rewriting the page.

**Architecture:** Keep business logic in the existing appointment page, extract only a pure summary helper that can be tested in isolation, and drive the visual refresh through shared theme tokens in `web/src/uni.scss`. Shared components consume CSS variables so the new palette propagates without forcing a page-by-page rewrite.

**Tech Stack:** uni-app, Vue 3, TypeScript, scoped SFC styles, existing deploy hook, Playwright verification

---

### Task 1: Add a testable appointment summary helper

**Files:**
- Create: `web/src/utils/appointment-calendar-summary.ts`
- Create: `web/scripts/test-appointment-calendar-summary.ts`
- Modify: `web/package.json`

- [ ] **Step 1: Write the failing test**

```ts
import { buildAppointmentCalendarSummary } from '../src/utils/appointment-calendar-summary'

const summary = buildAppointmentCalendarSummary([
  { ID: 1, status: 0, staff_id: 0 },
  { ID: 2, status: 1, staff_id: 8 },
  { ID: 3, status: 3, staff_id: 0 },
  { ID: 4, status: 4, staff_id: 0 },
])

if (summary.total !== 3) throw new Error(`expected total=3, got ${summary.total}`)
if (summary.pendingConfirm !== 1) throw new Error(`expected pendingConfirm=1, got ${summary.pendingConfirm}`)
if (summary.unassigned !== 2) throw new Error(`expected unassigned=2, got ${summary.unassigned}`)
if (summary.waitingCheckout !== 1) throw new Error(`expected waitingCheckout=1, got ${summary.waitingCheckout}`)
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
cd /Users/genglsh/workstation/cat/cat/web && npx tsc ./src/utils/appointment-calendar-summary.ts ./scripts/test-appointment-calendar-summary.ts --module commonjs --target es2020 --moduleResolution node --esModuleInterop --types node --skipLibCheck --outDir ./.tmp/appointment-calendar-summary-test && node ./.tmp/appointment-calendar-summary-test/scripts/test-appointment-calendar-summary.js
```

Expected: FAIL because `appointment-calendar-summary.ts` does not exist yet.

- [ ] **Step 3: Write minimal implementation**

```ts
export interface AppointmentCalendarSummary {
  total: number
  pendingConfirm: number
  unassigned: number
  waitingCheckout: number
}

export function buildAppointmentCalendarSummary(items: Array<{ status?: number; staff_id?: number | null }>): AppointmentCalendarSummary {
  return items.reduce((acc, item) => {
    if (item?.status !== 4) acc.total += 1
    if (item?.status === 0) acc.pendingConfirm += 1
    if (item?.status !== 4 && !item?.staff_id) acc.unassigned += 1
    if (item?.status === 3) acc.waitingCheckout += 1
    return acc
  }, {
    total: 0,
    pendingConfirm: 0,
    unassigned: 0,
    waitingCheckout: 0,
  })
}
```

- [ ] **Step 4: Run test to verify it passes**

Run the same command from step 2.

Expected: PASS and output `appointment calendar summary tests passed`.

- [ ] **Step 5: Add a reusable npm script**

```json
"test:appointment-calendar-summary": "rm -rf .tmp/appointment-calendar-summary-test && npx tsc ./src/utils/appointment-calendar-summary.ts ./scripts/test-appointment-calendar-summary.ts --module commonjs --target es2020 --moduleResolution node --esModuleInterop --types node --skipLibCheck --outDir ./.tmp/appointment-calendar-summary-test && node ./.tmp/appointment-calendar-summary-test/scripts/test-appointment-calendar-summary.js"
```

### Task 2: Add yellow theme tokens

**Files:**
- Modify: `web/src/uni.scss`

- [ ] **Step 1: Add receipt-inspired theme variables**

Add yellow-theme design tokens and map the uni primary palette to them:

```scss
$uni-color-primary: #d8a94f;
$uni-color-warning: #d2872c;
$uni-text-color: #3f3428;
$uni-text-color-grey: #8e7b62;
$uni-bg-color: #faf7f2;
$uni-bg-color-grey: #f6efe3;
$uni-border-color: #eadfcb;
```

- [ ] **Step 2: Add global CSS custom properties**

```scss
:root,
page,
body {
  --cat-color-primary: #d8a94f;
  --cat-color-primary-deep: #a07830;
  --cat-color-primary-soft: #fbf3df;
  --cat-color-page-bg: #faf7f2;
  --cat-color-card-bg: #fffaf3;
  --cat-color-card-bg-strong: #fff4df;
  --cat-color-text-main: #3f3428;
  --cat-color-text-muted: #8e7b62;
  --cat-color-border: #eadfcb;
  --cat-color-sidebar: #4c3a24;
  --cat-shadow-soft: 0 12rpx 28rpx rgba(116, 88, 38, 0.1);
}
```

### Task 3: Theme shared components

**Files:**
- Modify: `web/src/components/SideLayout.vue`
- Modify: `web/src/components/FilterPanel.vue`
- Modify: `web/src/components/CalendarPicker.vue`

- [ ] **Step 1: Swap hard-coded purple/cold neutrals for theme variables**
- [ ] **Step 2: Keep layout and interactions unchanged**
- [ ] **Step 3: Ensure active/highlight states still read clearly on mobile**

### Task 4: Upgrade appointment calendar first screen

**Files:**
- Modify: `web/src/pages/appointment/calendar.vue`

- [ ] **Step 1: Import the summary helper and expose a computed summary**
- [ ] **Step 2: Replace the old `stats-bar` with a summary card strip near the date navigation**
- [ ] **Step 3: Keep the existing pending panel, unassigned section, and staff grid behavior intact**
- [ ] **Step 4: Move the top area onto the yellow theme without touching fetch flows**

### Task 5: Verify and deploy

**Files:**
- Modify: none

- [ ] **Step 1: Run the focused summary test**

```bash
cd /Users/genglsh/workstation/cat/cat/web && npm run test:appointment-calendar-summary
```

- [ ] **Step 2: Run frontend type-check and note existing failures**

```bash
cd /Users/genglsh/workstation/cat/cat/web && npm run type-check
```

- [ ] **Step 3: Deploy web changes**

```bash
printf '{"tool_input":{"file_path":"/Users/genglsh/workstation/cat/cat/web/src/pages/appointment/calendar.vue"}}' | /Users/genglsh/workstation/cat/cat/.codex/hooks/deploy.sh
```

- [ ] **Step 4: Run Playwright verification**

Verify:
- appointment calendar page loads without console errors
- yellow theme is visible in navigation, filter panel, and calendar picker
- summary cards render above the staff grid
- pending panel still opens and closes
