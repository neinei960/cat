# Order Care Report Generation Design

## Context

The project already supports:

- paid order detail viewing on `web/src/pages/order/detail.vue`
- receipt image generation and save flows on the order detail page
- manual pet bath report uploads on `web/src/pages/pet/report.vue`
- bath report persistence through `PetBathReport`

The requested feature is to generate a structured care report image from a fixed template instead of uploading a finished image manually.

Reference assets supplied by the user:

- base template image: `base.jpg`
- rendered examples: `吧.jpg`, `b2.jpg`, `b3.jpg`

The key constraint is layout stability. The user explicitly rejected browser-side screenshot rendering because the final typography and checkbox alignment must remain fixed. The design therefore uses server-side template composition.

## Goal

Add a `生成报告` flow to completed order detail so staff can:

1. open a care report form from a paid order
2. choose one cat when the order contains multiple cats
3. upload and crop a portrait photo
4. fill report fields and checkboxes
5. submit the form to the backend
6. receive a server-rendered report image based on the fixed template
7. preview and save the image using the same save flow as receipts
8. automatically create a `pet_bath_reports` record for the selected cat

## Non-Goals

- No editable history for the structured report fields in this phase
- No redesign of the existing bath report management page
- No client-side `html2canvas` report rendering
- No generic template engine for multiple report styles
- No batch generation for multiple cats in one submission

## Chosen Approach

Use a hybrid flow:

- frontend handles report form UI, cat selection, portrait upload, and portrait cropping
- frontend uploads the cropped portrait with the existing upload API and sends the resulting URL to the backend
- backend composes the final image from the fixed template, fixed coordinates, and fixed font assets
- backend stores the rendered image in uploads and writes one `pet_bath_reports` row
- frontend previews the returned image and reuses the existing receipt save logic

This approach is preferred because:

- layout becomes device-independent
- typography and checkbox alignment are deterministic
- the final image is already a persistent URL, which fits the existing bath report data model
- frontend work stays limited to a modal workflow and does not need heavy canvas logic

## User Flow

### Entry

Surface:

- `web/src/pages/order/detail.vue`

Rules:

- show `生成报告` only when `order.status === 1`
- hide the action when the order has no resolvable cat

### Report creation

1. Staff opens `生成报告`
2. If the order maps to multiple cats, show a required cat picker first
3. Staff uploads a portrait photo
4. Staff crops the portrait using the same cropper interaction style already used elsewhere
5. Staff fills the report form
6. Staff submits
7. Backend returns the generated image URL and created bath report ID
8. Frontend opens preview mode with save actions

### Save

Reuse the current receipt save behavior:

- normal browsers: download or open image
- iPhone Safari: open dedicated image page for long-press save

## Data Rules

### Auto-filled fields

Populate these fields when entering the report form:

- pet name
- breed
- gender
- age
- care date
- care content

Field sources:

- pet name, breed, gender, age: selected order cat or related pet record
- care date: `order.pay_time`, fallback `order.CreatedAt`
- care content: service item names from the order, joined in display order

### Manual fields

Staff fills:

- weight
- suggested next care date
- portrait photo
- all checkbox groups
- all free-text notes

### Multi-cat rule

One submission generates one report for one cat only.

For multi-cat orders:

- user must select a single target cat before filling the form
- each cat requires a separate report submission

## Frontend Design

### Components

Add a local order-detail flow, preferably with a dedicated component:

- `web/src/components/order/OrderCareReportModal.vue`

Responsibilities:

- derive selectable cats from the current order
- render cat selector when needed
- host the portrait upload + crop flow
- render the report form
- submit the payload
- render preview mode from returned `image_url`
- reuse save-image behavior from receipt generation

### Cropper

Reuse the existing cropper utility:

- `web/src/utils/image-cropper.ts`

Flow:

- pick image
- crop to square
- upload cropped result through existing upload API
- store returned `portrait_url`

The frontend does not need to render the final template itself.

### Form model

Represent the report form as structured frontend state with:

- selected `pet_id`
- `portrait_url`
- `weight`
- `care_date`
- `next_care_date`
- `care_content`
- grouped checkbox values
- grouped note values

Checkbox groups should use stable codes instead of display text so backend rendering is deterministic.

## Backend Design

### API

Add one backend endpoint under the business order routes:

- `POST /b/orders/:id/care-report`

Request body should include:

- `pet_id`
- `portrait_url`
- `weight`
- `care_date`
- `next_care_date`
- `care_content`
- structured checkbox selections
- structured notes

Response should include:

- `image_url`
- `report_id`
- `bath_date`

### Validation

Backend must reject requests when:

- the order does not exist
- the order does not belong to the current shop
- the order is not paid
- the selected `pet_id` is not part of the order
- `portrait_url` is missing
- `next_care_date` is missing
- required template assets are unavailable

### Rendering service

Add a focused service layer for report composition, for example:

- `server/internal/service/order_care_report_service.go`

Responsibilities:

- resolve the target order and target cat
- normalize auto-filled values
- load the template base image
- load the fixed font file
- fetch the uploaded portrait image
- crop portrait into a circular frame
- draw text, lines, checkbox marks, and portrait onto the base template
- encode the result to JPG
- persist the final image into uploads
- create a `PetBathReport` row

### Template assets

Store and embed fixed assets in the server codebase, for example:

- base template image copied from user-provided `base.jpg`
- one bundled Chinese font file used for all text drawing

Use Go `embed` so deployment always includes the exact template and font version used to calibrate coordinates.

### Rendering engine

Use a backend drawing library built for fixed-coordinate image composition.

Recommended:

- `github.com/fogleman/gg`

Reasons:

- simple absolute-position drawing API
- supports text alignment and image compositing
- easier to maintain than raw `image/draw`

The report should always render at the original template size:

- `1279 x 1810`

### Coordinate model

Keep all drawing coordinates in one dedicated layout file, for example:

- `server/internal/service/order_care_report_layout.go`

The layout file should define:

- text anchors
- line baselines
- portrait frame bounds
- checkbox anchors
- note block bounds

This keeps the rendering logic deterministic and avoids hard-coded coordinates being scattered through business code.

## Persistence

After successful render:

1. write the generated JPG to the upload directory
2. expose it as `/uploads/...`
3. create one `PetBathReport` row with:
   - `shop_id`
   - `pet_id`
   - `image_url`
   - `bath_date`

This reuses the current bath report management flow without changing its data contract.

Structured form data will not be persisted in this phase.

Trade-off:

- reports can be viewed later as images
- reports cannot be reopened for structured editing later

This is acceptable for the first version because the immediate goal is stable template-based output.

## Text and Layout Rules

To keep the generated image visually stable:

- render with one bundled font only
- apply fixed font sizes per field group
- enforce max length or ellipsis on long fields
- constrain notes to fixed line counts
- keep checkbox labels static in the template; only draw checked marks

Recommended guardrails:

- pet name, breed, care content: single-line with truncation
- notes: max 1 to 2 lines depending on field width
- weight: normalized to a short decimal string
- dates: render in a single consistent format

## Error Handling

### Frontend

- disable submit while generating
- show inline validation for missing portrait or next-care date
- if generation fails, keep the filled form state so staff can retry
- if portrait upload fails, do not clear the crop result silently

### Backend

Return clear business errors for:

- invalid order state
- invalid cat selection
- invalid image source
- template rendering failure
- upload write failure
- bath report persistence failure

If image rendering succeeds but report persistence fails, do not return success. The output must be treated as one atomic creation flow, and any just-written generated image file should be deleted before returning the error.

## Verification Plan

### Backend tests

Add service-level tests for:

- paid order required
- selected cat must belong to order
- successful render creates upload output and bath report row
- generated image dimensions match template size
- long text is truncated instead of overflowing

### Frontend tests

Add focused regression checks for:

- multi-cat selection gating
- auto-fill mapping from order to form defaults
- modal state transitions between form and preview

### Manual UI verification

After deployment verify on H5:

- paid order shows `生成报告`
- portrait upload and crop succeed
- report submission returns preview image
- save-image flow works the same way as receipt save
- generated report appears in the selected cat’s bath report page

## Risks

- Template calibration will fail if the chosen font differs from the visual sample; the implementation must bundle the exact font asset used during calibration.
- Long Chinese text can push layout if truncation rules are too loose; limits must be explicit.
- If portrait uploads are heavily compressed before backend rendering, the circular portrait may look soft; frontend should upload a reasonably sized cropped image.
- Because structured report data is not stored, later template edits cannot regenerate historical reports from source fields.

## Follow-Up Options

Possible future extension, intentionally excluded from this phase:

- persist structured report payload alongside the rendered image
- support historical report re-edit and re-render
- support multiple report templates or brand themes
