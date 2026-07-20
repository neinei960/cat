# Order Care Report Visual Parity Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the H5 preview and backend-generated care-report JPG visually match the supplied real report for typography, dates, checkmarks, the care-content label, and long notes.

**Architecture:** Keep the backend as the source of truth for the saved JPG and preserve the fixed-template renderer. Add an embedded Bold font and deterministic layout helpers in the backend, then mirror the same display rules in the Vue preview without changing API or database contracts.

**Tech Stack:** Go, `fogleman/gg`, `x/image/font/opentype`, Vue 3, TypeScript, uni-app H5.

---

### Task 1: Lock the real-sample rules in failing tests

**Files:**
- Modify: `server/internal/service/order_care_report_service_test.go`
- Modify: `web/scripts/test-order-care-report-frontend.ts`

- [ ] **Step 1: Add backend expectations for date and checkmark geometry**

Add tests requiring ISO `2026-07-03` to display as `2026.7.3`, a thicker checkmark path, and unchanged scanned checkbox centers.

```go
func TestOrderCareReportDisplayDateUsesRealTemplateFormat(t *testing.T) {
    if got := formatOrderCareReportDisplayDate("2026-07-03"); got != "2026.7.3" {
        t.Fatalf("want 2026.7.3, got %q", got)
    }
}
```

- [ ] **Step 2: Add backend expectations for Bold text, label override, and two-line notes**

Add focused tests that require the Bold face to load, the old `Last care` label region to change, and an overlong note to wrap to at most two lines without crossing the next checkbox row.

```go
func TestWrapOrderCareReportNoteUsesAtMostTwoLines(t *testing.T) {
    lines := wrapOrderCareReportNote(face, longNote, 670, 2)
    if len(lines) != 2 { t.Fatalf("want 2 lines, got %d", len(lines)) }
}
```

- [ ] **Step 3: Add frontend source expectations**

Require `formatDisplayDate`, `care-content-label-override`, Bold field/note styles, larger checkmark dimensions, and two-line note styles.

- [ ] **Step 4: Run focused tests and confirm RED**

```bash
cd server
go test ./internal/service -run 'TestOrderCareReport(DisplayDate|Bold|Label|Checkmark|Note)' -count=1 -v

cd ../web
rm -rf .tmp/order-care-report-frontend-test
pnpm exec tsc ./scripts/test-order-care-report-frontend.ts --module commonjs --target es2020 --moduleResolution node --esModuleInterop --types node --skipLibCheck --outDir ./.tmp/order-care-report-frontend-test
node ./.tmp/order-care-report-frontend-test/test-order-care-report-frontend.js
```

Expected: the new assertions fail against the current Regular font, hyphenated dates, small checkmark, old label, and single-line notes.

### Task 2: Implement backend visual parity

**Files:**
- Create: `server/internal/service/assets/order-care-report/SourceHanSansSC-Bold.otf`
- Modify: `server/internal/service/order_care_report_service.go`
- Modify: `server/internal/service/order_care_report_layout.go`
- Test: `server/internal/service/order_care_report_service_test.go`

- [ ] **Step 1: Add the official Bold font and independent font cache**

Embed Regular and Bold separately and select the face explicitly.

```go
type orderCareReportFontWeight uint8
const (
    orderCareReportFontRegular orderCareReportFontWeight = iota
    orderCareReportFontBold
)
```

- [ ] **Step 2: Format display dates without changing input contracts**

```go
func formatOrderCareReportDisplayDate(value string) string {
    parsed, err := time.Parse("2006-01-02", strings.TrimSpace(value))
    if err != nil { return compactOrderCareReportText(value) }
    return fmt.Sprintf("%d.%d.%d", parsed.Year(), parsed.Month(), parsed.Day())
}
```

- [ ] **Step 3: Draw the care-content label override**

Cover only the old label rectangle with white, then draw `护理内容` and `Content of care` in Regular. Preserve the existing underline and value coordinate.

- [ ] **Step 4: Apply Bold sizes to user-entered fields**

Use Bold for primary values and notes, tune per-field sizes against the reference, and retain shrink-to-fit behavior.

- [ ] **Step 5: Match the authentic checkmark**

Increase stroke and offsets while retaining exact centers `406/569/732/895/1058` and the calibrated row centers.

- [ ] **Step 6: Support at most two Bold note lines**

Try one line from 24 px down to 18 px, then wrap to two centered lines and ellipsize only the second line. Keep both lines between the note baseline and next checkbox row.

- [ ] **Step 7: Run focused backend tests and confirm GREEN**

```bash
go test ./internal/service -run 'TestOrderCareReport|TestRenderOrderCareReport' -count=1 -v
```

Expected: all care-report layout and renderer tests pass.

### Task 3: Mirror the rules in the H5 preview

**Files:**
- Modify: `web/src/components/order/OrderCareReportStage.vue`
- Test: `web/scripts/test-order-care-report-frontend.ts`

- [ ] **Step 1: Add the care-content label overlay**

Add a fixed white overlay at the old label coordinates containing `护理内容` and `Content of care`, below the interactive hotspot layer.

- [ ] **Step 2: Mirror typography and display-date rules**

Set primary values and notes to `font-weight: 700`, update field sizes, and display ISO dates as unpadded dotted dates.

- [ ] **Step 3: Mirror the authentic checkmark path**

Update CSS width, height, border width, and offsets while keeping the backend anchor centers.

- [ ] **Step 4: Mirror two-line note layout**

Allow two centered lines with fixed line height and hidden overflow, using the same maximum line count and width rule as the backend.

- [ ] **Step 5: Run frontend test and H5 build**

```bash
rm -rf .tmp/order-care-report-frontend-test
pnpm exec tsc ./scripts/test-order-care-report-frontend.ts --module commonjs --target es2020 --moduleResolution node --esModuleInterop --types node --skipLibCheck --outDir ./.tmp/order-care-report-frontend-test
node ./.tmp/order-care-report-frontend-test/test-order-care-report-frontend.js
pnpm build:h5
```

Expected: source assertions pass and uni-app prints `DONE Build complete`.

### Task 4: Generate the real-reference artifact and verify regressions

**Files:**
- Modify: `server/internal/service/order_care_report_service_test.go` only when explicit real-reference fixture values are needed
- Create: `output/care-report-real-reference-v1.jpg`

- [ ] **Step 1: Generate a complete artifact**

Use the real-reference values, multiple checks, long notes, and a local cat image without persisting an order or bath report.

- [ ] **Step 2: Inspect the full image and critical crops**

Inspect main fields, checkmark rows, long notes, and the care-content label at original resolution. Correct mismatches before deployment.

- [ ] **Step 3: Run the full backend suite**

```bash
cd server
go test ./...
```

Expected: zero failures.

- [ ] **Step 4: Confirm no unintended diff**

Run `git diff --check` and inspect only renderer, preview, tests, font asset, and generated artifact changes.

### Task 5: Deploy and verify the remote build

**Files:**
- Deploy trigger: `server/internal/service/order_care_report_service.go`
- Deploy trigger: `web/src/components/order/OrderCareReportStage.vue`

- [ ] **Step 1: Deploy backend**

Run the repository hook. If its short health window rolls back a healthy slow-starting service, redeploy the same Linux binary manually, wait for migrations, and verify local/remote SHA-256 equality.

- [ ] **Step 2: Deploy frontend**

Run the hook for `OrderCareReportStage.vue` and confirm the remote order-detail JS/CSS contains the new date, label, Bold, checkmark, and note rules.

- [ ] **Step 3: Verify runtime health and browser loading**

Require remote `/api/v1/health` to return `status: ok`. Open the remote route in Playwright; if redirected to login, verify the public artifact URL and deployed bundle content.

- [ ] **Step 4: Upload the final artifact**

Upload `output/care-report-real-reference-v1.jpg` to `/opt/cat/web/`, compare local and remote hashes, and provide the public URL.
