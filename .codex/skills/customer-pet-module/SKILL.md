---
name: customer-pet-module
description: Diagnose, implement, or review the 客户与猫咪管理 domain in this repository. Use when the request involves 客户管理、客户标签、猫咪档案、洗护报告，或触及 customer/pet 在预约、订单、寄养、喂养中的选择、校验、展示、联动写入与关联查询。
---

# Customer Pet Module

Use this skill when the task touches the repo's 客户/猫咪主数据域. This domain is not limited to the management pages. It also drives selection, validation, and snapshots across appointment, order, boarding, and feeding flows.

## Required Superpower Order

1. If the request creates or changes behavior, invoke `brainstorming` first.
2. If the request is a bug or regression, invoke `systematic-debugging` first.
3. Before writing implementation code, invoke `test-driven-development`.
4. Before claiming the work is complete, invoke `verification-before-completion`.

Do not jump straight into page edits when the real issue is a cross-flow customer/pet invariant.

## Workflow

1. Identify the primary surface first.
   - Customer list/detail/edit/trash: `web/src/pages/customer/*`
   - Customer tag management: `web/src/pages/customer/tag-manage.vue`, `web/src/api/customer-tag.ts`
   - Pet list/edit/report: `web/src/pages/pet/*`, `web/src/api/pet.ts`, `web/src/api/pet-bath-report.ts`
   - Backend handlers: `server/internal/handler/customer.go`, `server/internal/handler/customer_tag.go`, `server/internal/handler/pet.go`
   - Backend services/repos/models: `server/internal/{service,repository,model}/customer*`, `server/internal/{service,repository,model}/pet*`, `server/internal/model/pet_bath_report.go`

2. Check whether the change is management-side or downstream flow-side.
   - Management-side requests usually affect customer/pet CRUD, search, tag binding, trash/restore, or report maintenance.
   - Downstream requests usually affect how customer/pet are selected, validated, preloaded, or snapshotted in appointment, order, boarding, or feeding flows.
   - Do not patch only the source page when the broken behavior appears in a downstream flow.

3. Confirm the data grain before changing behavior.
   - `Customer` is the owner/contact entity.
   - `CustomerTag` is a many-to-many label set on customer only.
   - `Pet` is an optional child of customer. Current schema allows `customer_id` to be `nil`.
   - `PetBathReport` is pet-level historical content, not a generic order or appointment note.

4. Read ownership and auto-create behavior before changing request shape.
   - Pet create/update accepts either `customer_id` or `owner_phone`.
   - If `owner_phone` is provided and no customer exists, backend auto-creates a customer.
   - That auto-created customer may be sparse. Do not assume nickname or full profile exists.

5. For cross-flow issues, inspect the consumer flow before editing shared entities.
   - Appointment uses customer/pet selection and validation.
   - Orders may preload customer and pet, and some flows support grouped pet data.
   - Boarding and feeding depend on customer/pet matching and downstream snapshots.
   - If a requested fix changes who can be selected together, verify both source entity data and consumer-side validation.

6. After changing `web/` or `server/`, deploy with the repo hook.
   - Required command:
   ```bash
   printf '{"tool_input":{"file_path":"/absolute/path/to/changed/file"}}' | /Users/genglsh/workstation/cat/cat/.codex/hooks/deploy.sh
   ```

7. For any UI-affecting change, run a browser check after deployment.
   - Verify the actual H5 rendering for customer list/detail/edit, pet list/edit/report, or the affected downstream picker/form.

## Rules That Matter

- Keep shop scoping intact on all customer, tag, pet, and bath-report queries/writes.
- Do not turn customer-pet association into a required relationship unless schema and downstream flows are updated together.
- Treat `owner_phone` auto-create behavior as part of current contract. If you change it, verify every pet create/update entry point.
- Customer soft delete and restore are live behavior. Verify trash/restore semantics before changing delete logic.
- Customer tag filtering belongs on customer queries; do not fake it only in frontend.
- Pet species defaults to `猫` in backend create flow. Preserve that unless the domain requirement truly changes.
- Do not overload `PetBathReport` into a generic visit record; it is ordered, pet-scoped historical content.
- Be careful with member-card-related customer fields. This skill includes member-card touchpoints only when they affect customer filtering, display, or downstream price linkage, not the full member-card module.

## Verification Checklist

- Backend change:
  - Verify request binding in affected handler
  - Verify preload/search behavior in the corresponding repository
  - Verify customer-pet ownership rules still hold
  - If the change affects downstream flows, verify the consumer handler/service too

- Frontend change:
  - Verify the exact customer or pet page in browser
  - Verify customer tag selection/filtering if tags are involved
  - Verify pet owner selection and owner display if association logic changed
  - Verify bath-report create/edit/reorder/delete if report UI changed

- Cross-layer change:
  - Verify `web/src/api/customer*.ts` or `web/src/api/pet*.ts` payload shape
  - Verify route registration in `server/internal/router/router.go` if endpoints changed
  - Verify affected appointment/order/boarding/feeding flow if they consume the changed fields
  - Verify deployed behavior after running the repo deploy hook

## Read This Reference

Load [customer-pet-module-map.md](references/customer-pet-module-map.md) when you need:
- customer/pet file map
- linked downstream flows
- current invariants and pitfalls
- where a requested change should actually land
