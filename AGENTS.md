# Codex Project Notes

## Project Structure

This repository is split into a uni-app H5 frontend and a Go/Gin backend.

- `web/`: Vue 3 + TypeScript + uni-app frontend.
- `web/src/pages/`: route pages, grouped by business domain such as `appointment`, `boarding`, `customer`, `feeding`, `order`, `pet`, `product`, `service`, `setting`, `shop`, and `staff`.
- `web/src/components/`: reusable UI components. Domain-specific shared components may live under subdirectories such as `web/src/components/order/`.
- `web/src/api/`: frontend API clients. Keep request/response mapping here instead of scattering raw request calls across pages.
- `web/src/store/`: Pinia stores and cross-page state.
- `web/src/types/`: shared TypeScript declarations.
- `web/src/utils/`: pure helpers and small browser/uni-app utilities.
- `server/`: Go backend.
- `server/cmd/`: executable entry points and maintenance commands.
- `server/config/`: configuration loading and examples.
- `server/internal/router/`: route registration only.
- `server/internal/handler/`: HTTP request binding, validation, and response shaping.
- `server/internal/service/`: business rules, workflows, calculations, and transaction orchestration.
- `server/internal/repository/`: persistence queries and database-facing operations.
- `server/internal/model/`: Gorm/domain models.
- `server/internal/middleware/`: Gin middleware.
- `server/pkg/`: reusable infrastructure packages such as auth, database, response, utils, and WeChat integration.

## Code Organization Rules

- Keep changes inside the owning domain module whenever possible. For example, boarding changes should primarily touch `boarding` pages, API files, handlers, services, repositories, and models.
- Follow the existing handler -> service -> repository layering. Do not put business workflows in handlers or database query construction in pages.
- Keep `router` changes limited to registering routes and middleware.
- Keep page components focused on UI state, user interaction, and orchestration. Move reusable formatting, parsing, and calculation logic into `utils`, `api`, store, or backend service code as appropriate.
- Prefer extending existing files and patterns over adding new abstractions. Add a new component, helper, service method, or repository method only when it has a clear owner and reduces real duplication or complexity.
- When a page grows too large, extract local child components or shared components without changing business behavior.
- Keep API contracts explicit. If backend response shapes change, update `web/src/api/*` and `web/src/types/*` together.
- Do not perform broad rewrites, formatting-only sweeps, or unrelated cleanup while implementing a focused change.
- Preserve existing user changes in the working tree. Never revert unrelated edits unless the user explicitly asks.

## Frontend Rules

- Use Vue 3 `<script setup lang="ts">` patterns already present in `web/src/pages/*`.
- Use existing uni-app primitives and feedback patterns such as `uni.showToast`, loading states, disabled states, and navigation helpers.
- Register new pages in `web/src/pages.json` when adding a route.
- Keep domain API calls inside `web/src/api/<domain>.ts`.
- Keep shared display rules in utilities such as `web/src/utils/*` instead of duplicating formatting in pages.
- For UI changes, verify real rendered behavior after deployment with Playwright.

## Backend Rules

- Keep request parsing and response formatting in `server/internal/handler`.
- Keep business rules in `server/internal/service`; add focused tests for service-level calculations and state transitions.
- Keep SQL/Gorm details in `server/internal/repository`; avoid leaking query details into handlers.
- Keep persistent data shape in `server/internal/model`.
- Use `server/pkg/response` helpers for HTTP responses where existing handlers do.
- When adding or changing routes, update `server/internal/router/router.go` and the relevant handler/service/repository together.

## Verification

- For frontend changes, run the narrowest relevant checks first, such as `pnpm type-check`, `pnpm build:h5`, or an existing `pnpm test:*` script under `web/`.
- For backend changes, run focused Go tests for touched packages, then broader `go test ./...` when the blast radius is unclear.
- For UI changes, deploy first using the hook below, then run a Playwright check against the deployed behavior.
- If verification cannot be run, report the exact command that was skipped or failed and why.

## Domain Skills

This repo includes Codex skills under `.codex/skills/`. Use the matching skill before touching these areas:

- `boarding-module`: boarding cabinets, holidays, discount policies, orders, dashboard, check-in, and check-out.
- `customer-pet-module`: customers, tags, pets, bath reports, and customer/pet selection in related workflows.
- `feeding-module`: feeding plans, visits, pricing, playtime, settlement, dashboard, and calendar.
- `product-module`: product categories, SKUs/specs, retail checkout, discounts, and commissions.
- `frontend-page-governor`, `frontend-governed-implementer`, `frontend-component-extractor`, `frontend-state-ux-normalizer`: existing H5 page cleanup, extraction, and state normalization.

## Auto Deploy

This repository already has a Claude deployment hook at:

- `/Users/genglsh/workstation/cat/cat/.claude/hooks/deploy.sh`

When editing this project with Codex, treat remote deployment as required work, not optional cleanup.

Rules:

- After any code change under `web/`, run the deploy hook so the H5 frontend is built and synced to the remote server.
- After any code change under `server/`, run the deploy hook so the Linux binary is rebuilt and restarted on the remote server.
- If both `web/` and `server/` change, deploy both.
- Do not skip deployment unless the user explicitly says not to deploy.
- For any change that affects UI, run a Playwright check after deployment to verify the actual rendered behavior.

Preferred command:

```bash
printf '{"tool_input":{"file_path":"/absolute/path/to/changed/file"}}' | /Users/genglsh/workstation/cat/cat/.codex/hooks/deploy.sh
```

Remote deployment target:

- host: `36.151.144.227`
- user: `root`
- web dir: `/opt/cat/web/`
- server dir: `/opt/cat/server/`
