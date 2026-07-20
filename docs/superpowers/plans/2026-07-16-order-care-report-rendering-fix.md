# Order Care Report Rendering Fix Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make every newly generated care report use the deterministic backend renderer so the final 1279 x 1810 image has stable text, portrait, note, and checkbox positions on iPhone, Android, and desktop browsers.

**Architecture:** Keep the current H5 report editor and interactive template preview, but make that preview non-authoritative. On submit, send the structured draft and uploaded portrait URL to the existing `POST /b/orders/:id/care-report` endpoint. The Go service validates order/pet ownership, draws the final image with the embedded template and font at fixed coordinates, saves the rendered JPG, creates the `PetBathReport` row, and returns the authoritative image URL for preview/save.

**Tech Stack:** uni-app H5, Vue 3, TypeScript, Gin, GORM, Go `embed`, `github.com/fogleman/gg`, existing upload API and mobile image-save helper.

---

## Scope And Acceptance Criteria

In scope:

- Restore the backend-rendering architecture already approved in `docs/superpowers/specs/2026-04-20-order-care-report-generation-design.md`.
- Preserve current report editing, cat selection, portrait crop/upload, preview, and save interactions.
- Preserve report-only edits to pet name, breed, gender, and age without writing those edits back to the pet profile.
- Remove `html2canvas` from the final care-report generation path.
- Add regression coverage for the frontend routing decision and backend fixed-coordinate output.
- Deploy backend first, then frontend, and verify the real remote H5 flow.

Out of scope:

- No automatic rewrite of existing report images.
- No structured historical report editing or regeneration.
- No new report template or report management redesign.
- No changes to customer/pet ownership rules.

Acceptance criteria:

- A newly generated report is returned as a `1279 x 1810` JPG from the backend.
- Generated text, notes, portrait, and checkmarks use the fixed coordinates in `order_care_report_layout.go` and do not depend on browser DPR, viewport width, scroll position, or system fonts.
- The modal calls `POST /b/orders/:id/care-report` exactly once per submission and no longer uploads an `html2canvas` result or separately calls `createPetBathReport`.
- The returned `image_url` is the image shown in the final preview and saved by the user.
- The saved filename uses `.jpg`, matching the backend image encoding and response URL.
- One successful submission creates exactly one `pet_bath_reports` row for the selected pet.
- A failed render or failed report-row write does not show success and does not leave a generated output file behind.
- Existing frontend validation and filled draft state remain available after a retryable failure.

## File Map

Workspace safety: the repository currently has unrelated uncommitted changes, including files that this fix will touch. Before each edit, inspect the current diff for that file and preserve all unrelated user changes. Do not use reset, checkout, mass formatting, or broad staging. Review only the care-report hunks before deployment.

Backend:

- Modify `server/internal/handler/order.go`: accept optional report-only display overrides and map them to the service input.
- Modify `server/internal/service/order_care_report_service.go`: render optional display overrides with database values as backward-compatible fallbacks.
- Modify `server/internal/service/order_care_report_service_test.go`: cover display overrides, all output anchors, dimensions, persistence, and cleanup.
- Modify `server/internal/handler/order_care_report_test.go`: cover the complete request/response contract.
- Keep `server/internal/service/order_care_report_layout.go` as the single backend coordinate source.
- Keep `server/internal/router/router.go` unchanged; the route already exists and its registration test remains part of verification.

Frontend:

- Create `web/src/api/order-care-report.ts`: typed wrapper for `POST /b/orders/:id/care-report`.
- Modify `web/src/utils/order-care-report.ts`: add the pure draft-to-request mapper.
- Modify `web/src/components/order/OrderCareReportModal.vue`: submit structured data to the backend and use the returned image URL.
- Modify `web/src/components/order/OrderCareReportStage.vue`: retain preview/editing only; remove image export and `html2canvas` usage.
- Modify `web/src/utils/web-image-save.ts`: make generated care-report filenames end in `.jpg` to match backend output.
- Modify `web/scripts/test-order-care-report-frontend.ts`: invert the old client-rendering assertions and enforce backend generation.
- Modify `web/scripts/test-order-care-report.ts`: cover request mapping and the JPG report filename.
- Delete `web/scripts/test-order-care-report-stage-image-ready.ts`: its image-readiness assertions only apply to the removed browser rasterization path.
- Modify `web/package.json` and `web/pnpm-lock.yaml`: remove `html2canvas`; the current repository-wide search shows the care-report stage is its only source usage.

## Exact API Contract

Use this frontend request type:

```ts
export interface OrderCareReportSectionInput {
  checks: string[]
  note: string
}

export interface CreateOrderCareReportRequest {
  pet_id: number
  portrait_url: string
  pet_name: string
  breed: string
  gender: string
  age: string
  weight: string
  care_date: string
  next_care_date: string
  care_content: string
  body_shape: string
  skin: OrderCareReportSectionInput
  hair: OrderCareReportSectionInput
  nails: OrderCareReportSectionInput
  eyes_face: OrderCareReportSectionInput
  ears: OrderCareReportSectionInput
  oral: OrderCareReportSectionInput
  anus: OrderCareReportSectionInput
}

export interface CreateOrderCareReportResponse {
  image_url: string
  report_id: number
  bath_date: string
}
```

The four added display fields are report-only overrides. The backend must not update `pets`. Use pointers in Go so omitted fields keep backward-compatible database fallbacks while explicitly supplied values, including an empty string, remain authoritative for the report image:

```go
PetName *string `json:"pet_name"`
Breed   *string `json:"breed"`
Gender  *string `json:"gender"`
Age     *string `json:"age"`
```

---

### Task 1: Lock The Backend Rendering Contract

**Files:**

- Modify: `server/internal/service/order_care_report_service_test.go`
- Modify: `server/internal/handler/order_care_report_test.go`
- Test: `server/internal/service/order_care_report_service_test.go`
- Test: `server/internal/handler/order_care_report_test.go`

- [ ] **Step 1: Write a failing service test for report-only display overrides**

Add a focused test that calls `buildOrderCareReportDrawData` with explicit values and proves they win over the pet profile without mutating the pet:

```go
func TestBuildOrderCareReportDrawDataUsesDisplayOverrides(t *testing.T) {
	pet := &model.Pet{Name: "数据库名字", Breed: "数据库品种", Gender: 1}
	petName := "报告名字"
	breed := "报告品种"
	gender := "MM"
	age := "2岁1月"

	data := buildOrderCareReportDrawData(pet, CreateOrderCareReportInput{
		PetName: &petName,
		Breed:   &breed,
		Gender:  &gender,
		Age:     &age,
	}, time.Date(2026, 7, 16, 0, 0, 0, 0, time.Local))

	if data.PetName != petName || data.Breed != breed || data.Gender != gender || data.Age != age {
		t.Fatalf("unexpected display data: %+v", data)
	}
	if pet.Name != "数据库名字" || pet.Breed != "数据库品种" || pet.Gender != 1 {
		t.Fatalf("report overrides must not mutate pet: %+v", pet)
	}
}
```

- [ ] **Step 2: Run the service test and confirm it fails before implementation**

Run:

```bash
cd /Users/genglsh/workstation/cat/cat/server
go test ./internal/service -run TestBuildOrderCareReportDrawDataUsesDisplayOverrides -count=1 -v
```

Expected: compilation fails because `CreateOrderCareReportInput` does not yet expose `PetName`, `Breed`, `Gender`, and `Age`.

- [ ] **Step 3: Expand the render regression fixture to cover every row**

Extend `TestRenderOrderCareReportDrawsFieldsOutsidePortraitArea` so its fixture fills all primary fields, all section notes, and one checkbox per row. Reuse `assertImageDiffersNear` at:

```go
orderCareReportPortraitFrame.CenterX, orderCareReportPortraitFrame.CenterY
orderCareReportPrimaryFieldBoxes["pet_name"]
orderCareReportPrimaryFieldBoxes["breed"]
orderCareReportPrimaryFieldBoxes["gender"]
orderCareReportPrimaryFieldBoxes["age"]
orderCareReportPrimaryFieldBoxes["care_content"]
orderCareReportPrimaryFieldBoxes["care_date"]
orderCareReportPrimaryFieldBoxes["next_care_date"]
orderCareReportPrimaryFieldBoxes["weight"]
orderCareReportBodyShapeAnchors["standard"]
each selected point in orderCareReportSectionLayouts
each populated NoteBox baseline in orderCareReportSectionLayouts
```

Also retain the exact bounds assertions:

```go
if rendered.Bounds().Dx() != orderCareReportWidth || rendered.Bounds().Dy() != orderCareReportHeight {
	t.Fatalf("unexpected report size: %v", rendered.Bounds())
}
```

- [ ] **Step 4: Expand the handler success-contract test**

Post the complete request body, including the four display overrides and every section. Assert HTTP 200, non-empty `image_url`, non-zero `report_id`, normalized `bath_date`, and exactly one persisted report row.

- [ ] **Step 5: Run the focused backend tests**

Run:

```bash
cd /Users/genglsh/workstation/cat/cat/server
go test ./internal/service -run 'OrderCareReport|RenderOrderCareReport|BuildOrderCareReport' -count=1 -v
go test ./internal/handler -run OrderCareReport -count=1 -v
```

Expected: the new override test fails; existing validation, cleanup, dimension, and route tests remain green.

### Task 2: Add Backward-Compatible Backend Display Overrides

**Files:**

- Modify: `server/internal/handler/order.go`
- Modify: `server/internal/service/order_care_report_service.go`
- Test: `server/internal/service/order_care_report_service_test.go`
- Test: `server/internal/handler/order_care_report_test.go`

- [ ] **Step 1: Add optional fields to the handler and service input types**

Add the pointer fields shown in the API contract to `createOrderCareReportReq` and `CreateOrderCareReportInput`, then pass them through in `OrderHandler.CreateCareReport`.

- [ ] **Step 2: Resolve overrides without changing pet data**

Add this helper in `order_care_report_service.go`:

```go
func resolveOrderCareReportDisplayValue(override *string, fallback string) string {
	if override != nil {
		return compactOrderCareReportText(*override)
	}
	return compactOrderCareReportText(fallback)
}
```

Update `buildOrderCareReportDrawData` so name, breed, gender, and age use explicit request values when provided; otherwise retain current database-derived behavior. Keep truncation in the existing drawing functions.

- [ ] **Step 3: Run focused tests and confirm they pass**

Run:

```bash
cd /Users/genglsh/workstation/cat/cat/server
go test ./internal/service -run 'OrderCareReport|RenderOrderCareReport|BuildOrderCareReport' -count=1
go test ./internal/handler -run OrderCareReport -count=1
```

Expected: PASS.

- [ ] **Step 4: Run formatting and the wider backend suite**

Run:

```bash
cd /Users/genglsh/workstation/cat/cat/server
gofmt -w internal/handler/order.go internal/service/order_care_report_service.go internal/service/order_care_report_service_test.go internal/handler/order_care_report_test.go
go test ./... -count=1
```

Expected: PASS. If an unrelated pre-existing package fails, record the exact package and failure without hiding it.

### Task 3: Restore The Typed Frontend Backend-Generation Path

**Files:**

- Create: `web/src/api/order-care-report.ts`
- Modify: `web/src/utils/order-care-report.ts`
- Modify: `web/src/utils/web-image-save.ts`
- Modify: `web/scripts/test-order-care-report-frontend.ts`
- Modify: `web/scripts/test-order-care-report.ts`

- [ ] **Step 1: Replace the old frontend-rendering assertions with failing backend-flow assertions**

Update `test-order-care-report-frontend.ts` to require:

```ts
assert(modalSource.includes("from '@/api/order-care-report'"), 'modal should import backend report generation api')
assert(modalSource.includes('createOrderCareReport('), 'modal should call backend report generation api')
assert(!modalSource.includes('createPetBathReport('), 'modal must not persist a client-rendered report separately')
assert(!modalSource.includes('exportPngBlob('), 'modal must not export the preview DOM')
assert(!stageSource.includes("from 'html2canvas'"), 'report stage must not depend on html2canvas')
assert(!stageSource.includes('html2canvas('), 'report stage must not rasterize browser DOM')
```

- [ ] **Step 2: Run the frontend regression script and confirm it fails**

Run:

```bash
cd /Users/genglsh/workstation/cat/cat/web
npx tsc ./scripts/test-order-care-report-frontend.ts --module commonjs --target es2020 --moduleResolution node --esModuleInterop --types node --skipLibCheck --outDir ./.tmp/order-care-report-frontend-test
node ./.tmp/order-care-report-frontend-test/test-order-care-report-frontend.js
```

Expected: FAIL because the modal currently imports `createPetBathReport` and the stage still uses `html2canvas`.

- [ ] **Step 3: Create the typed API wrapper**

Create `web/src/api/order-care-report.ts` using the exact request and response types above:

```ts
import { request } from './request'

export function createOrderCareReport(orderId: number, data: CreateOrderCareReportRequest) {
  return request<CreateOrderCareReportResponse>({
    url: `/b/orders/${orderId}/care-report`,
    method: 'POST',
    data,
  })
}
```

- [ ] **Step 4: Add a pure draft-to-request mapper**

Add `buildOrderCareReportPayload` to `web/src/utils/order-care-report.ts`:

```ts
export function buildOrderCareReportPayload(draft: OrderCareReportDraft): CreateOrderCareReportRequest {
  return {
    pet_id: draft.petId,
    portrait_url: draft.portraitUrl,
    pet_name: draft.petName,
    breed: draft.breed,
    gender: draft.gender,
    age: draft.age,
    weight: draft.weight,
    care_date: normalizeOrderCareReportDate(draft.careDate),
    next_care_date: normalizeOrderCareReportDate(draft.nextCareDate),
    care_content: draft.careContent,
    body_shape: draft.bodyShape,
    skin: draft.skin,
    hair: draft.hair,
    nails: draft.nails,
    eyes_face: draft.eyesFace,
    ears: draft.ears,
    oral: draft.oral,
    anus: draft.anus,
  }
}
```

Extend `web/scripts/test-order-care-report.ts` to assert exact field mapping, especially `eyesFace -> eyes_face` and dotted date normalization.

Update the existing filename assertion so the saved extension matches the backend encoding:

```ts
assertEqual(
  buildOrderCareReportFileName('NO167', '福福'),
  '护理报告_NO167_福福.jpg',
  'backend-rendered report filename'
)
```

- [ ] **Step 5: Run the pure helper test**

Run:

```bash
cd /Users/genglsh/workstation/cat/cat/web
npx tsc ./src/utils/order-care-report.ts ./src/api/order-care-report.ts ./scripts/test-order-care-report.ts --module commonjs --target es2020 --moduleResolution node --esModuleInterop --types node --skipLibCheck --outDir ./.tmp/order-care-report-test
node ./.tmp/order-care-report-test/scripts/test-order-care-report.js
```

Expected: PASS.

### Task 4: Switch The Modal To The Authoritative Backend Image

**Files:**

- Modify: `web/src/components/order/OrderCareReportModal.vue`
- Modify: `web/src/components/order/OrderCareReportStage.vue`
- Modify: `web/src/utils/web-image-save.ts`
- Delete: `web/scripts/test-order-care-report-stage-image-ready.ts`
- Modify: `web/package.json`
- Modify: `web/pnpm-lock.yaml`

- [ ] **Step 1: Replace the submit pipeline**

In `OrderCareReportModal.vue`:

- Remove `createPetBathReport`, the `CareReportStageExpose` type, and `stageRef` export dependency.
- Import `createOrderCareReport` and `buildOrderCareReportPayload`.
- Keep `validateDraft`, the session token, submit lock, loading state, and retry-safe draft state.
- Replace the client rasterize/upload/persist block with:

```ts
const response = await createOrderCareReport(
  Number(props.order.ID),
  buildOrderCareReportPayload(currentDraft)
)
if (modalSessionToken.value !== sessionToken || draft.value !== currentDraft || !props.visible) return
previewUrl.value = response.data.image_url
activeEditor.value = null
uni.showToast({ title: '护理报告已生成', icon: 'success' })
```

The existing cropped portrait upload remains unchanged; only the final report-image creation moves to the backend.

- [ ] **Step 2: Make the stage preview-only**

In `OrderCareReportStage.vue`:

- Remove the `html2canvas` import.
- Remove `exportPngBlob`, `waitForStageImagesReady`, and `defineExpose`.
- Remove `nextTick` from Vue imports if no longer used.
- Keep all template, display, hotspot, scale, and editing behavior unchanged.

- [ ] **Step 3: Align the saved filename with backend JPG output**

Change `buildOrderCareReportFileName` to return `.jpg`. Do not change receipt filenames or generic `saveImageByUrl` behavior.

- [ ] **Step 4: Remove obsolete tests and dependency**

Delete `test-order-care-report-stage-image-ready.ts`; `test-order-care-report-frontend.ts` now owns the preview-only contract. The current `package.json` does not register a script for the deleted file. Run:

```bash
cd /Users/genglsh/workstation/cat/cat/web
rg -n "html2canvas|exportPngBlob|waitForStageImagesReady" src scripts package.json
```

Expected after edits: no care-report generation references remain. Remove the now-unused dependency:

```bash
pnpm remove html2canvas
```

- [ ] **Step 5: Run targeted frontend tests**

Run:

```bash
cd /Users/genglsh/workstation/cat/cat/web
npx tsc ./scripts/test-order-care-report-frontend.ts --module commonjs --target es2020 --moduleResolution node --esModuleInterop --types node --skipLibCheck --outDir ./.tmp/order-care-report-frontend-test
node ./.tmp/order-care-report-frontend-test/test-order-care-report-frontend.js
npx tsc ./src/utils/order-care-report.ts ./src/api/order-care-report.ts ./scripts/test-order-care-report.ts --module commonjs --target es2020 --moduleResolution node --esModuleInterop --types node --skipLibCheck --outDir ./.tmp/order-care-report-test
node ./.tmp/order-care-report-test/scripts/test-order-care-report.js
pnpm build:h5
```

Expected: both scripts exit 0 and H5 build completes.

- [ ] **Step 6: Run type checking and record pre-existing failures separately**

Run:

```bash
cd /Users/genglsh/workstation/cat/cat/web
pnpm type-check
```

Expected: no new errors in the touched care-report files. The repository currently has unrelated historical type errors, so compare results against the pre-change baseline and report any remaining unrelated errors explicitly.

### Task 5: Deploy In A Backward-Compatible Order

**Files:**

- Deploy trigger: `server/internal/service/order_care_report_service.go`
- Deploy trigger: `web/src/components/order/OrderCareReportModal.vue`

- [ ] **Step 1: Capture a rollback snapshot**

Before deployment, record the current remote server binary timestamp and current web asset names:

```bash
ssh root@36.151.144.227 'stat -c "%n %s %y" /opt/cat/server/server /opt/cat/web/index.html'
curl -fsS http://36.151.144.227/ | rg 'assets/index-.*\.(js|css)'
```

- [ ] **Step 2: Deploy the backend first**

Run:

```bash
printf '{"tool_input":{"file_path":"/Users/genglsh/workstation/cat/cat/server/internal/service/order_care_report_service.go"}}' | /Users/genglsh/workstation/cat/cat/.codex/hooks/deploy.sh
```

Verify:

```bash
curl -fsS http://36.151.144.227/api/v1/health
ssh root@36.151.144.227 'systemctl is-active cat.service'
```

Expected: health succeeds and service is `active`.

- [ ] **Step 3: Deploy the frontend second**

Run:

```bash
printf '{"tool_input":{"file_path":"/Users/genglsh/workstation/cat/cat/web/src/components/order/OrderCareReportModal.vue"}}' | /Users/genglsh/workstation/cat/cat/.codex/hooks/deploy.sh
```

Expected: build and sync exit 0; the new order-detail chunk imports the typed care-report API and contains no `html2canvas` implementation.

- [ ] **Step 4: Recheck backend health and frontend asset availability**

Run:

```bash
curl -fsS http://36.151.144.227/api/v1/health
curl -fsS http://36.151.144.227/ | rg 'assets/index-.*\.js'
```

Expected: health remains successful and the page references the newly deployed asset hash.

### Task 6: Verify The Deployed Flow And Clean Up Test Data

**Files:**

- No source changes expected.
- Evidence output may be stored under `output/order-care-report-verification/` and must not be committed.

- [ ] **Step 1: Use a dedicated paid test order and pet**

Do not generate a verification report against an arbitrary customer pet. Before submitting in production, obtain a user-approved paid test order/pet, record its order ID and pet ID, and note the current `pet_bath_reports` count. Stop before submission if no approved verification order is available; local backend render tests and non-mutating browser checks may still proceed.

- [ ] **Step 2: Verify the H5 interaction in a real browser**

At a mobile viewport, verify:

```text
1. Open the paid order detail.
2. Open 生成报告.
3. Select the intended cat when multiple cats are present.
4. Upload and crop the known test portrait.
5. Fill report-only display overrides, dates, weight, body shape, every section, and notes.
6. Click 生成报告 once.
7. Confirm the final preview switches to the returned persisted image URL.
8. Confirm 保存图片 still uses the existing mobile save flow.
```

- [ ] **Step 3: Verify the authoritative network request**

Confirm one successful request:

```text
POST /api/v1/b/orders/<test-order-id>/care-report -> 200
```

Confirm there is no final report PNG upload request and no separate `POST /b/pets/:id/bath-reports` from the modal.

- [ ] **Step 4: Verify the generated file and persistence**

Fetch the returned image and inspect it:

```bash
curl -fsS "http://36.151.144.227<returned-image-url>" -o output/order-care-report-verification/generated.jpg
sips -g pixelWidth -g pixelHeight output/order-care-report-verification/generated.jpg
```

Expected:

```text
pixelWidth: 1279
pixelHeight: 1810
```

Visually compare the image against `web/src/assets/order-care-report-base.jpg` and confirm:

- portrait is centered inside the circular frame;
- every primary field sits on its intended underline;
- each selected checkmark sits inside the intended checkbox;
- each note sits on its own note line without overlapping adjacent rows;
- the final image is identical when viewed from narrow and wide mobile viewports because generation is server-side.

Query persistence and assert the count increased by exactly one for the selected pet. Confirm the stored `image_url` equals the API response.

- [ ] **Step 5: Clean up only the designated verification record**

After evidence is captured, delete the exact verification `PetBathReport` by its returned `report_id` through the existing report delete API. The current delete path only soft-deletes the database row and does not remove uploads, so also remove the exact generated report file and exact temporary portrait upload by their recorded `/uploads/<uuid>.<ext>` basenames. Verify both basenames came from this test run before deleting them. Do not delete or rewrite any pre-existing report.

- [ ] **Step 6: Keep historical reports unchanged**

Do not batch-regenerate existing images. `pet_bath_reports` currently stores only `image_url`, `bath_date`, and ordering metadata, not the original structured draft, so an automatic historical rewrite cannot reproduce the original selections and notes safely.

## Rollback

If the backend deployment fails before frontend deployment, the deploy hook automatically restores `/opt/cat/server/server-bak`; verify `cat.service` and keep the old frontend live.

If the frontend deployment fails after the backend succeeds, keep the new backend live: the added request fields are optional and backward-compatible, so the old frontend continues working. Restore `/opt/cat/web-prev` to `/opt/cat/web` or redeploy the previous frontend commit.

If the deployed endpoint returns bad images, stop new frontend submissions by rolling back only the frontend first. Do not delete any customer report automatically; identify the verification `report_id` and clean up only that row/file.

## Final Evidence Checklist

- [ ] Backend focused tests pass.
- [ ] Backend full test suite passes or unrelated failures are documented.
- [ ] Frontend care-report regression scripts pass.
- [ ] H5 build passes.
- [ ] No care-report usage of `html2canvas`, `exportPngBlob`, or separate `createPetBathReport` remains.
- [ ] Backend deployed before frontend.
- [ ] Remote health check passes after both deployments.
- [ ] One dedicated test report is `1279 x 1810` and visually aligned.
- [ ] Exactly one test report row was created, verified, and then cleaned up.
- [ ] Existing historical reports remain untouched.
