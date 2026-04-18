# Customer Record Collapse Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the customer detail member-card record area show only the latest record by default and expand/collapse the full list with an arrow trigger.

**Architecture:** Keep the change isolated to the customer detail page and a tiny regression script under `web/scripts/`. Reuse the existing `records` data source and local toggle state, update the derived display list, and swap the header action from text links to a compact arrow/count control.

**Tech Stack:** uni-app, Vue 3 `<script setup>`, TypeScript, existing repo deploy hook, Playwright MCP

---

### Task 1: Add a Failing Regression Check

**Files:**
- Create: `web/scripts/test-customer-detail-record-collapse.ts`
- Modify: `web/package.json`
- Test: `web/scripts/test-customer-detail-record-collapse.ts`

- [ ] **Step 1: Write the failing test**

```ts
import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'

const filePath = path.resolve(__dirname, '../../src/pages/customer/detail.vue')
const source = fs.readFileSync(filePath, 'utf8')

assert(
  source.includes('records.value.slice(0, 1)'),
  'collapsed customer detail records should keep only the latest record visible',
)

assert(
  !source.includes('records.value.slice(0, 3)'),
  'customer detail records should no longer default to three visible rows',
)

assert(
  source.includes(\"records-arrow\"),
  'customer detail records should render an arrow-style toggle control',
)

console.log('customer detail record collapse regression test passed')
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd web && npm run test:customer-detail-record-collapse`
Expected: FAIL because `customer/detail.vue` still uses `slice(0, 3)` and has no arrow toggle class

- [ ] **Step 3: Add the npm script**

```json
"test:customer-detail-record-collapse": "rm -rf .tmp/customer-detail-record-collapse-test && npx tsc ./scripts/test-customer-detail-record-collapse.ts --module commonjs --target es2020 --moduleResolution node --esModuleInterop --types node --skipLibCheck --outDir ./.tmp/customer-detail-record-collapse-test && node ./.tmp/customer-detail-record-collapse-test/test-customer-detail-record-collapse.js"
```

- [ ] **Step 4: Re-run the test and confirm the expected failure**

Run: `cd web && npm run test:customer-detail-record-collapse`
Expected: FAIL with the collapse-regression assertion message, not a compile or path error

### Task 2: Implement the Page Change

**Files:**
- Modify: `web/src/pages/customer/detail.vue`
- Test: `web/scripts/test-customer-detail-record-collapse.ts`

- [ ] **Step 1: Update the derived record list**

```ts
const showAllRecords = ref(false)
const hasExpandableRecords = computed(() => records.value.length > 1)
const displayRecords = computed(() => showAllRecords.value ? records.value : records.value.slice(0, 1))
```

- [ ] **Step 2: Replace the text toggle with an arrow/count trigger**

```vue
<view class="records-header">
  <text class="records-title">充值/消费记录</text>
  <view v-if="hasExpandableRecords" class="records-arrow" @click="showAllRecords = !showAllRecords">
    <text class="records-arrow-count">{{ records.length }}条</text>
    <text :class="['records-arrow-icon', showAllRecords ? 'open' : '']">⌄</text>
  </view>
</view>
```

- [ ] **Step 3: Add minimal styles for the arrow trigger**

```css
.records-arrow { display: inline-flex; align-items: center; gap: 8rpx; color: #4F46E5; }
.records-arrow-count { font-size: 22rpx; color: #9CA3AF; }
.records-arrow-icon { font-size: 26rpx; line-height: 1; transform: rotate(0deg); transition: transform 0.2s ease; }
.records-arrow-icon.open { transform: rotate(180deg); }
```

- [ ] **Step 4: Run the regression test to verify it passes**

Run: `cd web && npm run test:customer-detail-record-collapse`
Expected: PASS with `customer detail record collapse regression test passed`

### Task 3: Verify, Deploy, and Browser Check

**Files:**
- Modify: `web/src/pages/customer/detail.vue`
- Test: `web/scripts/test-customer-detail-record-collapse.ts`

- [ ] **Step 1: Run focused verification**

Run: `cd web && npm run test:customer-detail-record-collapse`
Expected: PASS

- [ ] **Step 2: Deploy the changed H5 page**

Run:

```bash
printf '{"tool_input":{"file_path":"/Users/genglsh/workstation/cat/cat/web/src/pages/customer/detail.vue"}}' | /Users/genglsh/workstation/cat/cat/.codex/hooks/deploy.sh
```

Expected: H5 build completes and remote sync succeeds

- [ ] **Step 3: Run browser verification after deployment**

Verify in Playwright on the customer detail page that:

```text
- multiple records: only one row visible before expanding
- arrow trigger is visible with total count
- clicking arrow reveals all rows
- clicking again collapses back to one row
- admin edit/delete controls still render on expanded rows where allowed
```

- [ ] **Step 4: Report verification evidence and residual risk**

```text
- mention regression test result
- mention deploy hook result
- mention Playwright interaction result
- mention any unrelated pre-existing page or auth problems encountered during browser verification
```
