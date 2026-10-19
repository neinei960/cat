# Product Module Map

## Scope

The current 商品管理 module covers:

- Product categories
- Product base info
- Product SKUs/specs
- Product search and list filtering
- Product selection inside retail/service order creation
- Product discount split in orders
- Product commission in orders

It does not cover:

- Inventory
- Inbound/outbound stock records
- Barcode or SKU code management
- Cost accounting
- Supplier management

## Core Files

### Frontend

- `web/src/pages/product/list.vue`
  - Product management entry
  - Category tab filter
  - Keyword search
  - Delete action

- `web/src/pages/product/edit.vue`
  - Product base info
  - Single-spec / multi-spec switch
  - SKU create/edit/delete
  - Brand suggestions

- `web/src/pages/product/category.vue`
  - Flat category management
  - Drag reorder
  - Create/edit/delete

- `web/src/api/product.ts`
  - Product and category request wrappers

- `web/src/pages/order/create.vue`
  - Product search and category filter in order creation
  - Sellable SKU filtering
  - Product cart and subtotal
  - Product discount display

- `web/src/pages/order/batch-create.vue`
  - Batch appointment order flow with added products

### Backend

- `server/internal/model/product.go`
  - `ProductCategory`
  - `Product`
  - `ProductSKU`

- `server/internal/handler/product.go`
  - Product CRUD
  - SKU request binding
  - Brand list endpoint

- `server/internal/handler/product_category.go`
  - Category CRUD
  - Category deletion guard

- `server/internal/repository/product_repo.go`
  - Product list search
  - SKU preload
  - SKU replace semantics

- `server/internal/service/product_service.go`
  - Thin wrapper over repository

- `server/internal/handler/order.go`
  - Draft/build order totals
  - Product discount split
  - Product commission

- `server/internal/service/order_service.go`
  - Member-card product discount resolution
  - Product totals in appointment/direct orders
  - Order kind classification
  - Retail grouping as `零售商品`

## Behavioral Notes

### Product and SKU

- Product price is derived from SKU rows.
- Single-spec mode still sends one SKU.
- Multi-spec mode requires full replacement of SKU rows.
- SKU `sellable` controls whether a spec can be selected in order UI.

### Search

Backend product keyword search matches:

- product name
- brand
- category name
- SKU spec name

If a search bug appears only for one of those fields, start in `server/internal/repository/product_repo.go`.

### Category

- Categories are flat and ordered by `sort_order ASC`.
- Category status exists, but product category management is mainly create/edit/reorder/delete.
- Deleting a category in use by products should fail.

### Orders

- Product totals are separate from service totals.
- Product discount comes from member card `product_discount_rate`.
- Service discount and product discount are intentionally separate.
- Product commission uses `staff.product_commission_rate`.
- A product-only order becomes `order_kind = product`.
- In grouped displays, product items are grouped under `零售商品`.

## Common Pitfalls

- Changing product edit UI without preserving the full `skus` payload can silently delete specs.
- Treating `status` and `sellable` as the same thing is wrong.
  - `status` is at product/category level.
  - `sellable` is at SKU level.
- Adding inventory-related UI without backend support will create fake behavior.
- Fixing discount display only in frontend can drift from backend settlement; check both sides.
- Changing category structure to two-level requires schema, UI, query, and order-filter changes; current module is flat.
