# Product Management — Phase 1 (4 Master Tables) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build full CRUD for the four product master tables — Brand, Category, Subcategory, UoM — following the existing Internal Users pattern, wired end-to-end (DB → Go API → Nuxt page).

**Architecture:** Vertical slice per entity. One new Go package `internal/modules/product` holds per-entity files (model+repo+handler each) + one `routes.go`. Frontend gets one page + one `useXxx` composable per entity under a new "Products" sidebar menu group. Server-side pagination + search + filter reused from the Internal Users implementation.

**Tech Stack:** Go + Gin + pgx (backend), PostgreSQL (uuid, audit cols, triggers), Nuxt 4 + @nuxt/ui (frontend), MinIO (images). Auth: admin-only management routes (reuse existing middleware).

## Global Constraints

- Table prefix `m_`, id `uuid` default `uuid_generate_v4()`, standard columns on every table: `is_active` (bool NOT NULL default true), `created_at`, `created_by` (uuid → m_internal_user), `modified_at`, `modified_by` (uuid → m_internal_user), trigger `update_modified_at_column` for `modified_at`.
- Next migration number: **005**.
- No automated test framework in this repo — verify each task with `go build ./...`, `curl` against `http://localhost:8082`, and browser check at `http://localhost:3005` (admin/admin123, or the in-app cookie login).
- Reuse existing frontend components: `SearchSort`, `TablePager`, `AppModal`, `EmptyState`, and the `useUsers.ts` composable shape (paginated `fetchPage`, server-side search/sort/filter, infinite scroll, read-only page badge).
- Reference implementation to mirror: `be/internal/modules/auth/{model,repository,handler,routes}.go` and `fe/app/pages/settings/internal-users.vue` + `fe/app/composables/useUsers.ts`.
- API list endpoints return `{ data: [...], total: N }` and accept `limit,offset,search,sort,desc` (+ entity-specific filters). Default order `created_at DESC`.

---

## File Structure

**Backend (new package `be/internal/modules/product/`):**
- `brand.go` — Brand model + repo + handler
- `category.go` — Category model + repo + handler
- `subcategory.go` — Subcategory model + repo + handler (joins category name)
- `uom.go` — UoM model + repo + handler
- `routes.go` — register all product-master routes (admin-only)
- `be/migrations/005_product_masters.sql` — the four tables

**Backend (modify):**
- `be/internal/app` router wiring (wherever auth routes are registered) — mount product routes
- `be/internal/modules/upload/` — add brand-logo & category-banner upload endpoints (or a generic image endpoint with a `folder` param)

**Frontend (new):**
- `fe/app/composables/useBrands.ts`, `useCategories.ts`, `useSubcategories.ts`, `useUoms.ts`
- `fe/app/pages/products/brands.vue`, `categories.vue`, `subcategories.vue`, `uom.vue`

**Frontend (modify):**
- `fe/app/components/AppSidebar.vue` — add "Products" menu group + submenu (Products, Brands, Categories, Subcategories, UoM)

---

## Task 1: Migration 005 — four master tables

**Files:**
- Create: `be/migrations/005_product_masters.sql`

- [ ] **Step 1: Write the migration**

```sql
-- Migration 005 — Product master tables: brand, category, subcategory, uom.
-- Mengikuti pola m_internal_user (uuid, audit cols, trigger modified_at).

CREATE TABLE IF NOT EXISTS m_brand (
    id           uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    name         text NOT NULL,
    description  text,
    logo         text,
    is_active    boolean NOT NULL DEFAULT true,
    created_at   timestamptz NOT NULL DEFAULT now(),
    created_by   uuid REFERENCES m_internal_user(id) ON DELETE SET NULL,
    modified_at  timestamptz NOT NULL DEFAULT now(),
    modified_by  uuid REFERENCES m_internal_user(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS m_category (
    id            uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    name          text NOT NULL,
    banner_image  text,
    is_active     boolean NOT NULL DEFAULT true,
    created_at    timestamptz NOT NULL DEFAULT now(),
    created_by    uuid REFERENCES m_internal_user(id) ON DELETE SET NULL,
    modified_at   timestamptz NOT NULL DEFAULT now(),
    modified_by   uuid REFERENCES m_internal_user(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS m_subcategory (
    id           uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    name         text NOT NULL,
    category_id  uuid NOT NULL REFERENCES m_category(id) ON DELETE RESTRICT,
    is_active    boolean NOT NULL DEFAULT true,
    created_at   timestamptz NOT NULL DEFAULT now(),
    created_by   uuid REFERENCES m_internal_user(id) ON DELETE SET NULL,
    modified_at  timestamptz NOT NULL DEFAULT now(),
    modified_by  uuid REFERENCES m_internal_user(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS m_uom (
    id           uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    name         text NOT NULL,
    description  text,
    is_active    boolean NOT NULL DEFAULT true,
    created_at   timestamptz NOT NULL DEFAULT now(),
    created_by   uuid REFERENCES m_internal_user(id) ON DELETE SET NULL,
    modified_at  timestamptz NOT NULL DEFAULT now(),
    modified_by  uuid REFERENCES m_internal_user(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_subcategory_category ON m_subcategory (category_id);

-- Trigger modified_at untuk tiap tabel (fungsi update_modified_at_column sudah ada dari migration 002).
DO $$
DECLARE t text;
BEGIN
  FOREACH t IN ARRAY ARRAY['m_brand','m_category','m_subcategory','m_uom'] LOOP
    EXECUTE format('DROP TRIGGER IF EXISTS update_%1$s_modified_at ON %1$s;', t);
    EXECUTE format('CREATE TRIGGER update_%1$s_modified_at BEFORE UPDATE ON %1$s FOR EACH ROW EXECUTE FUNCTION update_modified_at_column();', t);
  END LOOP;
END $$;
```

- [ ] **Step 2: Apply the migration**

Run (use the DB creds from `be/.env`, port 5435, user/db `zalio_erp`):
```bash
cd be && PGPASSWORD="$(grep -E '^DB_PASSWORD' .env | cut -d= -f2-)" \
  psql -h localhost -p 5435 -U zalio_erp -d zalio_erp -v ON_ERROR_STOP=1 -f migrations/005_product_masters.sql
```
Expected: `CREATE TABLE` ×4, `CREATE INDEX`, `DO`.

- [ ] **Step 3: Verify tables exist**

Run: `PGPASSWORD=... psql -h localhost -p 5435 -U zalio_erp -d zalio_erp -c "\dt m_brand|m_category|m_subcategory|m_uom" -c "\d m_subcategory"`
Expected: 4 tables listed; `m_subcategory` shows `category_id` FK to `m_category`.

- [ ] **Step 4: Commit**

```bash
git add be/migrations/005_product_masters.sql
git commit -m "feat(product): migration 005 — brand/category/subcategory/uom master tables"
```

---

## Task 2: Backend — Brand CRUD (package `product`)

**Files:**
- Create: `be/internal/modules/product/brand.go`
- Create: `be/internal/modules/product/routes.go` (initial, brand only; extended in later tasks)
- Modify: router wiring file where `auth` routes are registered (mirror how auth is mounted) — mount `product` routes with admin-only middleware.

**Interfaces:**
- Produces: `product.RegisterRoutes(rg *gin.RouterGroup, pool *pgxpool.Pool, tm *token.Manager)` (or match the exact signature auth uses — inspect `auth/routes.go` first and mirror it).
- Brand JSON shape: `{ id, name, description, logo, is_active, created_at }`.
- List endpoint: `GET /api/v1/brands?limit&offset&search&sort&desc` → `{ data, total }`.
- `POST /api/v1/brands`, `PUT /api/v1/brands/:id`, `PATCH /api/v1/brands/:id/toggle-active`.

- [ ] **Step 1: Inspect the template**

Read `be/internal/modules/auth/{model.go,repository.go,handler.go,routes.go}` and the file that calls `auth.RegisterRoutes` (grep for `RegisterRoutes` / `auth.`). Mirror the exact signatures, middleware, and error-handling style (23505 → 409, per-field errors where relevant).

- [ ] **Step 2: Write `brand.go`**

Implement, mirroring the auth pattern:
- `type Brand struct` with json tags: `ID string` (uuid::text), `Name`, `Description`, `Logo`, `IsActive bool`, `CreatedAt time.Time`.
- `type brandRepo struct{ pool *pgxpool.Pool }` with:
  - `const brandCols = "id::text, name, COALESCE(description,'') AS description, COALESCE(logo,'') AS logo, is_active, created_at"`
  - `ListPaged(ctx, limit, offset int, search, sort string, desc bool) ([]Brand, int, error)` — search on `name`; whitelist sort: `name`; default `created_at DESC`; secondary `, id`. Same `$1='' OR name ILIKE '%'||$1||'%'` filter pattern as auth.
  - `Create(ctx, name, description, logo string) (*Brand, error)`
  - `Update(ctx, id, name, description, logo string) (*Brand, error)`
  - `ToggleActive(ctx, id string, active bool) (*Brand, error)` (`WHERE id=$2::uuid`)
- Handler methods: `ListBrands`, `CreateBrand` (bind `{name required, description, logo}`), `UpdateBrand`, `ToggleBrandActive` — same request/response shape and status codes as auth's user handlers.

- [ ] **Step 3: Write `routes.go` and mount it**

- `func RegisterRoutes(...)` grouping `/brands` under the admin-only middleware (mirror auth's admin group). Register the 4 brand routes.
- In the router wiring file, add the `product.RegisterRoutes(...)` call next to `auth.RegisterRoutes(...)`.

- [ ] **Step 4: Build**

Run: `cd be && go build ./...`
Expected: no errors.

- [ ] **Step 5: Restart backend & verify via curl**

```bash
pkill -f "go run ./cmd/api"; sleep 1; (cd be && nohup go run ./cmd/api > /tmp/zalio_be.log 2>&1 &) ; sleep 5
BASE=http://localhost:8082/api/v1
TOK=$(curl -s -X POST $BASE/auth/login -H 'Content-Type: application/json' -d '{"username":"admin","password":"admin123"}' | python3 -c 'import sys,json;print(json.load(sys.stdin)["token"])')
curl -s -X POST $BASE/brands -H "Authorization: Bearer $TOK" -H 'Content-Type: application/json' -d '{"name":"Nestle","description":"FMCG"}'
curl -s "$BASE/brands?limit=8&offset=0" -H "Authorization: Bearer $TOK"
```
Expected: create returns the brand JSON with an id; list returns `{"data":[{...Nestle...}],"total":1}`.

- [ ] **Step 6: Commit**

```bash
git add be/internal/modules/product/brand.go be/internal/modules/product/routes.go <router-file>
git commit -m "feat(product): brand CRUD API (paginated list, create, update, toggle)"
```

---

## Task 3: Backend — Category CRUD

**Files:**
- Create: `be/internal/modules/product/category.go`
- Modify: `be/internal/modules/product/routes.go` (add `/categories` routes)

**Interfaces:**
- Category JSON: `{ id, name, banner_image, is_active, created_at }`.
- Endpoints: `GET/POST /api/v1/categories`, `PUT /categories/:id`, `PATCH /categories/:id/toggle-active`.

- [ ] **Step 1: Write `category.go`** — identical structure to `brand.go` (Task 2), fields `name` (required), `banner_image`. `catCols = "id::text, name, COALESCE(banner_image,'') AS banner_image, is_active, created_at"`. Search on `name`, sort whitelist `name`.
- [ ] **Step 2: Add `/categories` routes in `routes.go`.**
- [ ] **Step 3: Build** — `cd be && go build ./...` → no errors.
- [ ] **Step 4: Verify via curl** — create `{"name":"Beverages"}`, list returns it with `total:1`.
- [ ] **Step 5: Commit** — `git commit -m "feat(product): category CRUD API"`

---

## Task 4: Backend — Subcategory CRUD (with category join)

**Files:**
- Create: `be/internal/modules/product/subcategory.go`
- Modify: `be/internal/modules/product/routes.go` (add `/subcategories`)

**Interfaces:**
- Subcategory JSON: `{ id, name, category_id, category_name, is_active, created_at }` — includes joined `category_name` for table display.
- Endpoints: `GET/POST /api/v1/subcategories`, `PUT`, `PATCH toggle-active`. Optional filter `?category_id=` for the cascading dropdown in the product form later.

- [ ] **Step 1: Write `subcategory.go`** — like brand, but:
  - `subCols = "s.id::text, s.name, s.category_id::text, c.name AS category_name, s.is_active, s.created_at"` with `FROM m_subcategory s JOIN m_category c ON c.id = s.category_id`.
  - `Create(ctx, name, categoryID string)`, `Update(ctx, id, name, categoryID string)`.
  - `ListPaged` supports optional `categoryID` filter param (`($X='' OR s.category_id = $X::uuid)`), search on `s.name`, sort whitelist `name`.
  - Handler `CreateSubcategory` binds `{name required, category_id required}`.
- [ ] **Step 2: Add `/subcategories` routes.**
- [ ] **Step 3: Build** → no errors.
- [ ] **Step 4: Verify via curl** — create with the Beverages category id, list shows `category_name:"Beverages"`.
- [ ] **Step 5: Commit** — `git commit -m "feat(product): subcategory CRUD API (with category join)"`

---

## Task 5: Backend — UoM CRUD

**Files:**
- Create: `be/internal/modules/product/uom.go`
- Modify: `be/internal/modules/product/routes.go` (add `/uoms`)

**Interfaces:**
- UoM JSON: `{ id, name, description, is_active, created_at }`. Endpoints `GET/POST /api/v1/uoms`, `PUT`, `PATCH toggle-active`.

- [ ] **Step 1: Write `uom.go`** — identical to `brand.go` with fields `name` (required), `description`.
- [ ] **Step 2: Add `/uoms` routes.**
- [ ] **Step 3: Build** → no errors.
- [ ] **Step 4: Verify via curl** — create `{"name":"Kilogram","description":"kg"}`, list returns it.
- [ ] **Step 5: Commit** — `git commit -m "feat(product): uom CRUD API"`

---

## Task 6: Backend — image upload for brand logo & category banner

**Files:**
- Modify: `be/internal/modules/upload/{handler.go,routes.go}`

**Interfaces:**
- Add `POST /api/v1/upload/image?folder=<brand_logo|category_banner>` (or two endpoints `/upload/brand-logo`, `/upload/category-banner`). Same validation as profile-image (image only, ≤2MB), returns `{ path }`. Mirror the existing `UploadProfileImage`.

- [ ] **Step 1: Inspect `upload/handler.go`** — reuse the profile-image logic; parameterize the MinIO folder.
- [ ] **Step 2: Implement the generic/extra endpoint(s)** writing to folders `brand_logo/` and `category_banner/`.
- [ ] **Step 3: Build** → no errors.
- [ ] **Step 4: Verify via curl** — upload a small PNG, get a `path`; GET `/files/<path>` → 200.
- [ ] **Step 5: Commit** — `git commit -m "feat(upload): brand-logo & category-banner image endpoints"`

---

## Task 7: Frontend — "Products" sidebar menu group

**Files:**
- Modify: `fe/app/components/AppSidebar.vue`

**Interfaces:**
- Adds a "Products" parent menu (flyout submenu like "Settings") with items in this order: `Products` (`/products`), `Brands` (`/products/brands`), `Categories` (`/products/categories`), `Subcategories` (`/products/subcategories`), `UoM` (`/products/uom`). Products (`/products`) page itself is built in Phase 2 — for now the link can point to `/products/brands` or a placeholder; the four master links are live after this phase.

- [ ] **Step 1: Read `AppSidebar.vue`** — mirror the existing "Settings" submenu/flyout mechanism (`showSettings`, `settingsItems`, hover flyout).
- [ ] **Step 2: Add a `productItems` array + a "Products" menu entry** reusing the same flyout markup/handlers. Icon e.g. `i-lucide-package`.
- [ ] **Step 3: Browser-verify** — start dev server, hover "Products", the 5 submenu items appear in order.
- [ ] **Step 4: Commit** — `git commit -m "feat(product): Products sidebar menu group"`

---

## Task 8: Frontend — Brands page + composable

**Files:**
- Create: `fe/app/composables/useBrands.ts`
- Create: `fe/app/pages/products/brands.vue`

**Interfaces:**
- `useBrands()` mirrors `useUsers()`: exposes `items` (ref), `total` (ref), `fetchPage(opts)`, `createBrand(body)`, `updateBrand(id,body)`, `toggleActive(row)`, plus `uploadImage(file)` for the logo. `ManagedBrand` interface `{ id, name, description, logo, is_active, created_at }`.
- `brands.vue` mirrors `internal-users.vue`: table (columns: Logo, Name, Description, Status) + infinite-scroll pagination + search + create/edit modal (fields: Name required, Description, Logo upload) + active toggle. Breadcrumbs `Products › Brands`.

- [ ] **Step 1: Read `useUsers.ts` and `internal-users.vue`** as the template.
- [ ] **Step 2: Write `useBrands.ts`** — copy the `useUsers` shape, swap endpoint to `/api/v1/brands` and fields to brand's. Keep the `fetchPage({offset,limit,search,sort,desc,append})` + `total` pattern.
- [ ] **Step 3: Write `brands.vue`** — copy `internal-users.vue`, strip user-specific fields, keep: search/sort, infinite scroll (`scrollEl`, `onScroll`, `updateCurrentPage`, `reload`, `loadMore`, `fillViewport`), read-only `TablePager`, create/edit `AppModal` with Name/Description/Logo (reuse the avatar-upload pattern for the logo). Table columns: Logo, Name, Description, Status(toggle).
- [ ] **Step 4: Browser-verify** — open `/products/brands`, add a brand with a logo, see it in the table; toggle active; edit it.
- [ ] **Step 5: Commit** — `git commit -m "feat(product): Brands page (CRUD + logo upload)"`

---

## Task 9: Frontend — Categories page + composable

**Files:**
- Create: `fe/app/composables/useCategories.ts`
- Create: `fe/app/pages/products/categories.vue`

**Interfaces:**
- `useCategories()` like `useBrands()`; `ManagedCategory { id, name, banner_image, is_active, created_at }`; endpoint `/api/v1/categories`; `uploadImage` → category-banner endpoint.
- `categories.vue` mirrors `brands.vue`: table columns Banner, Name, Status; modal fields Name (required) + Banner image upload. Breadcrumbs `Products › Categories`.

- [ ] **Step 1: Copy `useBrands.ts` → `useCategories.ts`**, swap endpoint/fields (name, banner_image).
- [ ] **Step 2: Copy `brands.vue` → `categories.vue`**, swap fields (Name + Banner upload), columns (Banner, Name, Status).
- [ ] **Step 3: Browser-verify** — add/edit a category with a banner; toggle; list works.
- [ ] **Step 4: Commit** — `git commit -m "feat(product): Categories page (CRUD + banner upload)"`

---

## Task 10: Frontend — Subcategories page + composable

**Files:**
- Create: `fe/app/composables/useSubcategories.ts`
- Create: `fe/app/pages/products/subcategories.vue`

**Interfaces:**
- `useSubcategories()`; `ManagedSubcategory { id, name, category_id, category_name, is_active, created_at }`; endpoint `/api/v1/subcategories`. Also expose a way to fetch categories for the dropdown (call `useCategories().fetchPage` with a big limit, or add a lightweight `fetchAllCategories()`).
- `subcategories.vue`: table columns Name, Category (shows `category_name`), Status; modal fields Name (required) + Category dropdown (options from categories). Breadcrumbs `Products › Subcategories`.

- [ ] **Step 1: Write `useSubcategories.ts`** (mirror useBrands; add category-name field).
- [ ] **Step 2: Write `subcategories.vue`** — mirror brands.vue; in the modal add a `<select>` Category populated by fetching categories (limit=1000, is_active) on mount. Table shows `category_name`.
- [ ] **Step 3: Browser-verify** — create a subcategory under "Beverages", table shows the category name; edit; toggle.
- [ ] **Step 4: Commit** — `git commit -m "feat(product): Subcategories page (CRUD + category dropdown)"`

---

## Task 11: Frontend — UoM page + composable

**Files:**
- Create: `fe/app/composables/useUoms.ts`
- Create: `fe/app/pages/products/uom.vue`

**Interfaces:**
- `useUoms()`; `ManagedUom { id, name, description, is_active, created_at }`; endpoint `/api/v1/uoms`.
- `uom.vue`: table columns Name, Description, Status; modal fields Name (required) + Description. Breadcrumbs `Products › UoM`.

- [ ] **Step 1: Copy `useBrands.ts` → `useUoms.ts`** (fields name, description; no image).
- [ ] **Step 2: Copy `brands.vue` → `uom.vue`** (drop logo; fields Name + Description; columns Name, Description, Status).
- [ ] **Step 3: Browser-verify** — add "Kilogram"/"kg", edit, toggle, list/search work.
- [ ] **Step 4: Commit** — `git commit -m "feat(product): UoM page (CRUD)"`

---

## Task 12: End-to-end smoke check & push

- [ ] **Step 1: Full click-through** — for each of the 4 pages: create, edit, toggle active, search, and scroll (add >8 rows to confirm infinite scroll). Confirm no console errors (ignore Vite HMR noise).
- [ ] **Step 2: Confirm menu order** — Products sidebar shows Products, Brands, Categories, Subcategories, UoM in that order.
- [ ] **Step 3: Push** — `git push` (remote `origin` = rizkinkbr-oss).

---

## Self-Review (done at plan-writing time)

- **Spec coverage:** Phase-1 scope (§7 Tahap 1) = the 4 masters — Tasks 1–11 cover DB (T1), backend CRUD (T2–T5), image upload for logo/banner (T6), menu (T7), and the 4 pages (T8–T11). Product/variant tables (§4) are intentionally deferred to Phase 2/3 plans. ✅
- **Placeholder scan:** No TBD/TODO. The one forward-reference — `/products` (Products page) link in the sidebar — is explicitly noted as Phase 2 with an interim target. ✅
- **Type consistency:** JSON shapes and composable method names (`fetchPage`, `total`, `createX`, `updateX`, `toggleActive`) are consistent across tasks and match the existing `useUsers.ts` contract. Repo `ListPaged` signature consistent across entities. ✅
- **Note:** Exact `RegisterRoutes` signature and router-wiring file are resolved by inspecting `auth` in Task 2 Step 1 before writing, since the precise wiring file wasn't captured in the spec.
