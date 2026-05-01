# Order Care Report Frontend Rendering Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace backend-drawn order care report generation with a frontend-rendered 1:1 report stage that exports a PNG, uploads it, and writes a pet bath report record.

**Architecture:** Keep the existing order detail entry and draft-building logic, but swap the submit path from `POST /b/orders/:id/care-report` to a client-side rendering pipeline: render a hidden report stage, capture it with `html2canvas`, upload the PNG through the existing upload API, then persist the resulting image URL with the existing pet bath report create API. Preserve current preview/save interactions so generated reports still open in the same modal and reuse the current H5 save behavior.

**Tech Stack:** uni-app H5, Vue 3 + TypeScript, `html2canvas`, existing upload API, existing pet bath report API

---

### Task 1: Rewire the Care Report Data Contract

**Files:**
- Modify: `/Users/genglsh/workstation/cat/cat/web/src/utils/order-care-report.ts`
- Modify: `/Users/genglsh/workstation/cat/cat/web/src/api/order-care-report.ts`

- [ ] Remove the backend request payload type and replace it with frontend-only helpers for date normalization, field display, and stage rendering text.
- [ ] Keep `buildOrderCareReportDraft()` and pet option discovery intact so order-derived defaults do not regress.
- [ ] Delete the obsolete `createOrderCareReport()` API helper once the modal no longer depends on `/b/orders/:id/care-report`.

### Task 2: Capture and Persist a Frontend-Rendered Report

**Files:**
- Modify: `/Users/genglsh/workstation/cat/cat/web/src/components/order/OrderCareReportModal.vue`
- Modify: `/Users/genglsh/workstation/cat/cat/web/src/api/pet-bath-report.ts`
- Reuse: `/Users/genglsh/workstation/cat/cat/web/src/api/upload.ts`
- Reuse: `/Users/genglsh/workstation/cat/cat/web/src/utils/web-image-save.ts`

- [ ] Add a hidden or isolated report stage in the modal that mirrors the final report layout instead of relying on backend drawing.
- [ ] On submit, render that stage with `html2canvas`, turn the canvas into a blob/file, upload it via `uploadH5File()`, then call `createPetBathReport(petId, imageUrl, bathDate)`.
- [ ] Update preview state to use the uploaded image URL returned from the persistence flow so the user sees the actual saved report.
- [ ] Preserve the existing portrait crop/upload flow and current Safari-friendly save behavior.

### Task 3: Build the 1:1 Frontend Report Stage

**Files:**
- Modify: `/Users/genglsh/workstation/cat/cat/web/src/components/order/OrderCareReportModal.vue`

- [ ] Add a dedicated report stage section sized to the template aspect ratio and driven entirely by the draft fields.
- [ ] Render the base template as a background layer and position text, portrait, body-shape selection, checkmarks, and notes with CSS so alignment can be tuned visually in the browser.
- [ ] Keep the editing form below or alongside the stage without introducing a new page or route.

### Task 4: Add Regression Coverage for the Frontend Flow

**Files:**
- Modify or create: `/Users/genglsh/workstation/cat/cat/web/scripts/test-order-care-report.ts`

- [ ] Add or extend a script that asserts the modal no longer imports or calls the backend order care report API.
- [ ] Cover the new success path at the unit level where practical: capture pipeline produces an uploaded URL and persists it through the pet bath report API.
- [ ] Keep the script runnable with the existing local TypeScript test pattern used in this repo.

### Task 5: Verify, Deploy, and Browser-Check

**Files:**
- Modify if needed after verification: `/Users/genglsh/workstation/cat/cat/web/src/components/order/OrderCareReportModal.vue`

- [ ] Run the targeted frontend regression script.
- [ ] Run `pnpm build:h5` in `/Users/genglsh/workstation/cat/cat/web`.
- [ ] Deploy the `web/` change with `/Users/genglsh/workstation/cat/cat/.codex/hooks/deploy.sh`.
- [ ] Use the browser to generate a real report from an order, confirm the generated PNG visually aligns with the stage, confirm the image persists into the pet bath report list, and confirm “保存图片” still works on H5.
