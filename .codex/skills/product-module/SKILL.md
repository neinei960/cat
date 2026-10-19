---
name: product-module
description: Diagnose, implement, or review the 商品管理 module in this repository. Use when the request involves 商品分类、商品编辑、SKU/规格、商品开单、零售商品结算、商品折扣、商品提成，或触及 `web/src/pages/product/*`, `web/src/api/product.ts`, `server/internal/{handler,service,repository,model}/product*`, 以及订单里的商品相关逻辑。
---

# Product Module

Use this skill when the task touches the repo's 商品管理 flow. This module is not only CRUD. It also affects retail order creation, member-card product discounts, and staff product commission.

## Workflow

1. Identify the layer first.
   - Product list / filter / delete: `web/src/pages/product/list.vue`
   - Product create/edit and SKU maintenance: `web/src/pages/product/edit.vue`
   - Product category management and ordering: `web/src/pages/product/category.vue`
   - API surface: `web/src/api/product.ts`
   - Backend handlers: `server/internal/handler/product.go`, `server/internal/handler/product_category.go`
   - Persistence/query logic: `server/internal/repository/product_repo.go`
   - Data model: `server/internal/model/product.go`
   - Order-side product selection and settlement: `web/src/pages/order/create.vue`, `web/src/pages/order/batch-create.vue`, `server/internal/handler/order.go`, `server/internal/service/order_service.go`

2. Confirm the data grain before changing behavior.
   - `Product` is the product shell.
   - `ProductSKU` is the actual price-bearing spec row.
   - Single-spec products still persist as one SKU with empty `spec_name`.
   - Product category is flat. Do not implement parent/child assumptions unless the schema is changed first.

3. Check whether the task is management-side or order-side.
   - Management-side requests usually affect product/category CRUD, search, ordering, or page UX.
   - Order-side requests usually affect selectable products, sellable SKU filtering, product subtotal, discount split, order kind, or commission.
   - Do not patch only the admin UI when the real bug is in order creation or settlement.

4. Read backend settlement logic first for any price-related change.
   - Product discount rate comes from member card product discount, not the general customer service discount.
   - Product totals and discount amounts are recalculated in order logic.
   - Staff product commission is separate from service commission.

5. Verify SKU semantics before editing UI.
   - The update path replaces all SKUs; it is not incremental.
   - If the UI edits specs, confirm the submitted `skus` array is complete.
   - Be careful when changing single-spec and multi-spec switching; this can discard spec data by design.

6. After changing `web/` or `server/`, deploy with the repo hook.
   - Required command:
   ```bash
   printf '{"tool_input":{"file_path":"/absolute/path/to/changed/file"}}' | /Users/genglsh/workstation/cat/cat/.codex/hooks/deploy.sh
   ```

7. For any UI-affecting change, run a browser check after deployment.
   - Verify the real H5 rendering for product list, product edit, category sort, or order-create product picker.
   - If the change affects retail ordering, verify product selection, SKU selection, subtotal, and discount display in browser.

## Rules That Matter

- Treat SKU price as the source of truth for product selling price.
- Do not add logic that depends on product inventory; the current module has no stock model or stock deduction flow.
- Keep shop scoping intact on product/category queries and writes.
- Preserve the distinction between `status` on product/category and `sellable` on SKU.
- Order-side product selection must filter out inactive products and SKUs with `sellable = false`.
- Product category deletion must keep the current guard: categories in use by products cannot be deleted.
- Do not assume SKU IDs are stable across updates. Current backend hard-deletes and recreates all SKUs on replace.
- When changing order totals, verify `ProductTotal`, `ProductDiscountAmount`, `DiscountAmount`, `PayAmount`, `Commission`, and `order_kind` together.

## Verification Checklist

- Backend change:
  - Read `server/internal/model/product.go`
  - Verify query/update behavior in `server/internal/repository/product_repo.go`
  - Verify handler request/response shape in `server/internal/handler/product.go`
  - If order totals changed, verify `server/internal/handler/order.go` and `server/internal/service/order_service.go`

- Frontend change:
  - Verify product list search, category tabs, and delete flow
  - Verify product edit create/edit flow for both single-spec and multi-spec
  - Verify category create/edit/delete/reorder
  - Verify order-create product picker, SKU picker, cart quantity updates, and discount summary

- Cross-layer change:
  - Verify `web/src/api/product.ts` request shape
  - Verify `server/internal/router/router.go` registration if API paths changed
  - Verify deployed behavior in browser after running the repo deploy hook

## Read This Reference

Load [product-module-map.md](references/product-module-map.md) when you need:
- product file map
- settlement and commission touchpoints
- current constraints and pitfalls
- where a requested change should actually land
