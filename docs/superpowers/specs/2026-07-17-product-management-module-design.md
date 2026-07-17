# Spec — Modul Product Management

**Tanggal:** 2026-07-17
**Status:** Disetujui untuk implementasi (menunggu review akhir)
**Modul:** Product Management (master data produk untuk Zalio ERP back-office)

---

## 1. Tujuan & Ruang Lingkup

Membangun modul **master data produk** sebagai sumber kebenaran operasional (bukan tampilan web commerce). Terdiri dari 6 tabel: 4 master pendukung + produk induk + varian produk.

**Termasuk (in scope):**
- 4 master: Brand, Category, Subcategory, UoM.
- Product master (`m_product`) + Product variant (`m_product_variant`), mendukung produk **single** & **variant**.
- CRUD lengkap tiap entitas mengikuti pola halaman **Internal Users** yang sudah ada.

**Tidak termasuk (out of scope, modul lain nanti):**
- **Product Catalog** (daftar produk tayang di web commerce) → tabel terpisah `ec_product_catalog` yang menunjuk (FK) ke master ini + field khusus toko (slug/SEO, harga tampil/promo, is_published, dll). Dibangun di modul ecommerce.
- **Chart of Accounts / Accounting** → kolom `coa_*` di produk untuk sekarang hanya **placeholder** (lihat §4).
- **Inventory / stok aktual** → modul terpisah (karena itu `min_stock`/stok tidak dimasukkan ke master produk).

---

## 2. Konvensi (mengikuti pola `m_internal_user`)

Semua tabel master memakai prefix `m_` dan **kolom standar** berikut:

| Kolom | Tipe | Ket. |
|---|---|---|
| `id` | uuid PK default `uuid_generate_v4()` | |
| `is_active` | boolean NOT NULL default true | |
| `created_at` | timestamptz NOT NULL default now() | |
| `created_by` | uuid → `m_internal_user(id)` ON DELETE SET NULL | |
| `modified_at` | timestamptz NOT NULL default now() | via trigger `update_modified_at_column` |
| `modified_by` | uuid → `m_internal_user(id)` ON DELETE SET NULL | |

- Migration berikutnya mulai dari **`005`**.
- Gambar (`logo`, `banner_image`, `main_image`, `image_1..3`) disimpan sebagai **path teks**, di-upload lewat MinIO (perluasan modul `upload` yang ada — endpoint/folder baru per entitas).

---

## 3. Tabel Master (kolom bisnis; kolom standar §2 otomatis ada)

### 3.1 `m_brand`
| Kolom | Tipe | Wajib |
|---|---|---|
| `name` | text | ✅ |
| `description` | text | — |
| `logo` | text (path gambar) | — |

### 3.2 `m_category`
| Kolom | Tipe | Wajib |
|---|---|---|
| `name` | text | ✅ |
| `banner_image` | text (path gambar) | — |

### 3.3 `m_subcategory`
| Kolom | Tipe | Wajib |
|---|---|---|
| `name` | text | ✅ |
| `category_id` | uuid → `m_category(id)` | ✅ |

### 3.4 `m_uom`
| Kolom | Tipe | Wajib |
|---|---|---|
| `name` | text (mis. "Kilogram") | ✅ |
| `description` | text | — |

---

## 4. Tabel Produk

### 4.1 `m_product` (induk / info bersama semua varian)

| Kolom | Tipe | Wajib | Ket. |
|---|---|---|---|
| `product_name` | text | ✅ | nama produk |
| `product_type` | text CHECK IN ('single','variant') default 'single' | ✅ | |
| `brand_id` | uuid → `m_brand(id)` | — | tidak semua produk bermerek |
| `subcategory_id` | uuid → `m_subcategory(id)` | ✅ | kategori ikut otomatis dari subkategori |
| `description` | text | — | |
| `ingredients` | text | — | di UI di sebelah kanan description |
| `is_perishable` | boolean default false | ✅ | punya kadaluarsa atau tidak |
| `uom_1` | uuid → `m_uom(id)` | ✅ | satuan dasar |
| `uom_2` | uuid → `m_uom(id)` | — | satuan ke-2 |
| `ratio_2` | numeric | — | 1 `uom_2` = `ratio_2` × `uom_1` |
| `uom_3` | uuid → `m_uom(id)` | — | satuan ke-3 |
| `ratio_3` | numeric | — | 1 `uom_3` = `ratio_3` × `uom_1` |
| `selling_uom` | uuid → `m_uom(id)` | — | satuan jual, dipilih dari uom_1/2/3 |
| `variant_name_1` | text | — | nama sumbu varian 1 (mis. "Warna") |
| `variant_name_2` | text | — | nama sumbu varian 2 (mis. "Ukuran") |
| `coa_inventory` | uuid (placeholder) | — | akun akuntansi — lihat catatan COA |
| `coa_sales` | uuid (placeholder) | — | |
| `coa_sales_return` | uuid (placeholder) | — | |
| `coa_sales_discount` | uuid (placeholder) | — | |
| `coa_good_in_transit` | uuid (placeholder) | — | |
| `coa_cogs` | uuid (placeholder) | — | |
| `coa_purchase_return` | uuid (placeholder) | — | |
| `coa_unbilled_goods` | uuid (placeholder) | — | |

**Catatan COA:** 8 kolom `coa_*` dibuat **nullable, tanpa FK constraint dulu** (tipe uuid). Saat modul Akuntansi/Chart-of-Accounts dibuat, ditambahkan FK ke tabel akun. Untuk sekarang boleh kosong.

### 4.2 `m_product_variant` (unit yang benar-benar dijual)

| Kolom | Tipe | Wajib | Ket. |
|---|---|---|---|
| `product_id` | uuid → `m_product(id)` ON DELETE CASCADE | ✅ | |
| `sku` | text UNIQUE | ✅ | kode produk |
| `barcode` | text (unik kalau diisi) | — | |
| `variant_value_1` | text | — | nilai sumbu 1 (mis. "Merah") |
| `variant_value_2` | text | — | nilai sumbu 2 (mis. "S") |
| `def_selling_price` | numeric(15,2) default 0 | — | harga jual default |
| `def_purchase_price` | numeric(15,2) default 0 | — | harga beli default |
| `cogs_unit` | numeric(15,2) default 0 | — | HPP per unit (diupdate terus) |
| `length_cm` | numeric | — | |
| `width_cm` | numeric | — | |
| `height_cm` | numeric | — | |
| `weight_gr` | numeric | — | |
| `main_image` | text (path) | — | |
| `image_1` | text (path) | — | |
| `image_2` | text (path) | — | |
| `image_3` | text (path) | — | |

---

## 5. Konsep Single & Variant

Model **2 tabel** (ala Shopify/Odoo), satu model untuk keduanya:

- `m_product.product_type` = `single` atau `variant`.
- **Single** → tepat **1 baris** `m_product_variant`. `variant_name_1/2` (induk) & `variant_value_1/2` (varian) **kosong**. Baris varian tetap menyimpan data jual (sku, harga, gambar, dimensi, cogs). Di UI terlihat seperti satu form biasa.
- **Variant** → banyak baris `m_product_variant`, dibedakan oleh **2 sumbu** varian (mis. Warna × Ukuran). Induk menetapkan `variant_name_1/2`; tiap varian mengisi `variant_value_1/2`, `sku`, `barcode`, harga, gambar, dimensi masing-masing.

---

## 6. UI / Navigasi

**Grup menu baru "Products"** di sidebar (flyout submenu, seperti "Settings"). Urutan submenu:

| # | Submenu | Halaman |
|---|---|---|
| 1 | **Products** (pertama) | `m_product` + varian |
| 2 | Brands | `m_brand` |
| 3 | Categories | `m_category` |
| 4 | Subcategories | `m_subcategory` |
| 5 | UoM | `m_uom` |

**Cakupan tiap halaman:** CRUD lengkap mengikuti pola **Internal Users** — tabel (search + sort + filter + infinite-scroll pagination) + modal tambah/edit + toggle aktif. Pakai ulang komponen: `SearchSort`, `TablePager`, `AppModal`, `EmptyState`, pola composable `useXxx.ts`.

- **Brands / Categories / UoM:** form sederhana.
- **Subcategories:** form ada dropdown pilih Category.
- **Products:** form terbesar — kemungkinan dibagi **beberapa seksi/tab** (Info umum, UoM & harga, Gambar, Akuntansi/COA, Varian). Category → Subcategory bertingkat (pilih kategori dulu untuk memfilter subkategori). Untuk `variant`, ada sub-tabel varian.

---

## 7. Urutan Pembuatan (bertahap)

Urutan **build** beda dengan urutan **menu** (produk butuh master sebagai FK):

- **Tahap 1 — 4 Master:** `m_brand` → `m_category` → `m_subcategory` → `m_uom` (CRUD sederhana; migration 005).
- **Tahap 2 — Produk single:** `m_product` + `m_product_variant`, dukung `product_type='single'` (migration 006).
- **Tahap 3 — Produk variant:** UI banyak varian + 2 sumbu opsi.

COA tetap placeholder sampai modul Akuntansi ada.

---

## 8. Sisi Backend (ringkas)

- Modul Go baru (pola vertical-slice seperti `auth`): `model.go`, `repository.go`, `handler.go`, `routes.go` untuk tiap entitas (atau dikelompokkan dalam satu package `product`). Ditentukan detailnya di rencana implementasi.
- Perluas modul `upload` untuk gambar brand/category/produk (folder terpisah per jenis).
- Endpoint mengikuti pola pagination server-side + filter yang sudah dipakai di Internal Users.
