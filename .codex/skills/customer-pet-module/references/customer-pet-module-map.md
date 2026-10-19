# Customer Pet Module Map

## Scope

The current 客户/猫咪 domain covers:

- Customer CRUD, search, detail, trash, restore
- Customer tags and tag-based filtering
- Pet CRUD, search, owner association
- Pet bath reports
- Customer/pet lookup and validation in appointment flows
- Customer/pet linkage in order flows
- Customer/pet matching inside boarding and feeding flows

It does not fully own:

- Member card lifecycle and recharge rules
- Service pricing logic itself
- Boarding pricing rules
- Feeding pricing rules
- General order settlement beyond customer/pet linkage

## Core Files

### Frontend

- `web/src/pages/customer/list.vue`
  - Customer management entry
  - Keyword search
  - Tag filtering
  - Customer quick actions

- `web/src/pages/customer/detail.vue`
  - Customer detail
  - Customer tag editing
  - Customer-linked pets display
  - Related profile updates

- `web/src/pages/customer/edit.vue`
  - Customer create/edit
  - Tag selection
  - Address and door-code maintenance

- `web/src/pages/customer/tag-manage.vue`
  - Customer tag CRUD

- `web/src/pages/customer/trash.vue`
  - Soft-deleted customer list
  - Restore flow

- `web/src/pages/pet/list.vue`
  - Pet management entry
  - Keyword and tag filtering

- `web/src/pages/pet/edit.vue`
  - Pet create/edit
  - Owner binding through customer or phone

- `web/src/pages/pet/report.vue`
  - Pet bath report CRUD and reorder

- `web/src/api/customer.ts`
  - Customer CRUD
  - Deleted list
  - Restore
  - Customer-linked pet query

- `web/src/api/customer-tag.ts`
  - Customer tag CRUD

- `web/src/api/pet.ts`
  - Pet CRUD and search

- `web/src/api/pet-bath-report.ts`
  - Pet bath report CRUD and reorder

### Backend

- `server/internal/model/customer.go`
  - `Customer`
  - Customer member-card and address fields
  - Pets and customer tags relations

- `server/internal/model/customer_tag.go`
  - `CustomerTag`

- `server/internal/model/customer_tag_relation.go`
  - many-to-many join table

- `server/internal/model/pet.go`
  - `Pet`
  - optional `customer_id`
  - pet profile fields used by downstream flows

- `server/internal/model/pet_bath_report.go`
  - `PetBathReport`

- `server/internal/handler/customer.go`
  - Customer CRUD
  - Deleted list
  - Restore
  - Customer-linked pet list

- `server/internal/handler/customer_tag.go`
  - Customer tag CRUD

- `server/internal/handler/pet.go`
  - Pet CRUD
  - `owner_phone` fallback and auto-create behavior

- `server/internal/repository/customer_repo.go`
  - Customer list/search
  - Tag filtering
  - Deleted/restore

- `server/internal/repository/customer_tag_repo.go`
  - Tag list and relation count

- `server/internal/repository/pet_repo.go`
  - Pet list/search
  - owner preload and customer lookup

- `server/internal/repository/pet_bath_report_repo.go`
  - report CRUD and sort order

- `server/internal/router/router.go`
  - API registration
  - downstream routes that also consume customer/pet data

## Downstream Consumers

- `web/src/pages/appointment/create.vue`
  - customer and pet selection
  - customer/pet consistency checks

- `web/src/pages/order/create.vue`
  - customer and pet selection
  - order-side rendering of selected pets

- `web/src/pages/boarding/create.vue`
  - customer and pet selection for boarding orders

- `web/src/pages/feeding/*`
  - customer and pet linkage for feeding plans

- `server/internal/handler/appointment.go`
  - appointment request binding using customer/pet

- `server/internal/handler/order.go`
  - order request and list logic with customer/pet fields

- `server/internal/handler/boarding.go`
  - boarding order create/update using customer/pet IDs

- `server/internal/handler/feeding.go`
  - plan listing and operations keyed by customer/pet data

## Behavioral Notes

### Customer

- Customer search and list support keyword plus optional member-card-template and customer-tag filtering.
- Customer delete is soft delete, not hard delete.
- Customer restore is exposed in management UI and API.
- Customer tags are stored in a many-to-many relation and also appear in response payloads.

### Pet

- Pet can exist without an owner.
- Backend defaults pet species to `猫` on create when omitted.
- Pet create/update accepts `customer_id` directly, or `owner_phone` as a lookup-and-attach shortcut.
- If `owner_phone` does not match an existing customer, backend auto-creates one with sparse data.

### Bath Reports

- Bath reports are queried by pet.
- Report order is persisted and reorder is part of the API surface.
- Treat them as pet-profile history, not order records.

## Common Pitfalls

- Fixing customer or pet display in one page while ignoring downstream pickers causes drift.
- Assuming every pet must have a customer breaks current create and import flows.
- Treating customer tags as a plain text field is wrong for current management UI; the active UI uses relational tags.
- Changing phone-based owner binding without checking auto-create paths will break pet creation from lightweight forms.
- Patching only frontend filtering for customer tags or deleted data can diverge from backend pagination totals.
- Mixing member-card logic into customer CRUD can accidentally change unrelated pricing behavior.
