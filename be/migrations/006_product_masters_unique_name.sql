-- Migration 006 — cegah nama duplikat (case-insensitive) di master produk.
-- Unik pakai lower(name) supaya "Nestle" == "nestle".
-- Subcategory: unik GLOBAL (bukan per kategori).

CREATE UNIQUE INDEX IF NOT EXISTS uq_brand_name_lower ON m_brand (lower(name));
CREATE UNIQUE INDEX IF NOT EXISTS uq_category_name_lower ON m_category (lower(name));
CREATE UNIQUE INDEX IF NOT EXISTS uq_uom_name_lower ON m_uom (lower(name));
CREATE UNIQUE INDEX IF NOT EXISTS uq_subcategory_name_lower ON m_subcategory (lower(name));
