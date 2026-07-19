-- Migration 008 — Produk: m_product (induk) + m_product_variant (unit jual).
-- Mendukung produk single & variant. COA & country_of_origin = placeholder.

CREATE TABLE IF NOT EXISTS m_product (
    id                 uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    product_name       text NOT NULL,
    product_type       text NOT NULL DEFAULT 'single' CHECK (product_type IN ('single', 'variant')),
    brand_id           uuid REFERENCES m_brand(id) ON DELETE RESTRICT,
    subcategory_id     uuid NOT NULL REFERENCES m_subcategory(id) ON DELETE RESTRICT,
    country_of_origin  text,                      -- placeholder; nanti dari country list
    description        text,
    ingredients        text,
    is_perishable      boolean NOT NULL DEFAULT false,
    uom_1              uuid NOT NULL REFERENCES m_uom(id) ON DELETE RESTRICT,
    uom_2              uuid REFERENCES m_uom(id) ON DELETE RESTRICT,
    ratio_2            numeric,
    uom_3              uuid REFERENCES m_uom(id) ON DELETE RESTRICT,
    ratio_3            numeric,
    selling_uom        uuid REFERENCES m_uom(id) ON DELETE RESTRICT,
    variant_name_1     text,
    variant_name_2     text,
    -- 8 akun COA (placeholder; FK ke chart-of-accounts ditambah saat modul akuntansi ada)
    coa_inventory        uuid,
    coa_sales            uuid,
    coa_sales_return     uuid,
    coa_sales_discount   uuid,
    coa_good_in_transit  uuid,
    coa_cogs             uuid,
    coa_purchase_return  uuid,
    coa_unbilled_goods   uuid,
    is_active          boolean NOT NULL DEFAULT true,
    created_at         timestamptz NOT NULL DEFAULT now(),
    created_by         uuid REFERENCES m_internal_user(id) ON DELETE SET NULL,
    modified_at        timestamptz NOT NULL DEFAULT now(),
    modified_by        uuid REFERENCES m_internal_user(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS m_product_variant (
    id                  uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    product_id          uuid NOT NULL REFERENCES m_product(id) ON DELETE CASCADE,
    sku                 text NOT NULL,
    barcode             text,
    variant_value_1     text,
    variant_value_2     text,
    def_selling_price   numeric(15,2) NOT NULL DEFAULT 0,
    def_purchase_price  numeric(15,2) NOT NULL DEFAULT 0,
    cogs_unit           numeric(15,2) NOT NULL DEFAULT 0,
    length_cm           numeric,
    width_cm            numeric,
    height_cm           numeric,
    weight_gr           numeric,
    main_image          text,
    image_1             text,
    image_2             text,
    image_3             text,
    is_active           boolean NOT NULL DEFAULT true,
    created_at          timestamptz NOT NULL DEFAULT now(),
    created_by          uuid REFERENCES m_internal_user(id) ON DELETE SET NULL,
    modified_at         timestamptz NOT NULL DEFAULT now(),
    modified_by         uuid REFERENCES m_internal_user(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_product_subcategory ON m_product (subcategory_id);
CREATE INDEX IF NOT EXISTS idx_product_brand ON m_product (brand_id);
CREATE INDEX IF NOT EXISTS idx_variant_product ON m_product_variant (product_id);

-- SKU unik (case-insensitive). Barcode unik hanya kalau diisi.
CREATE UNIQUE INDEX IF NOT EXISTS uq_variant_sku_lower ON m_product_variant (lower(sku));
CREATE UNIQUE INDEX IF NOT EXISTS uq_variant_barcode ON m_product_variant (barcode) WHERE barcode IS NOT NULL AND barcode <> '';

-- Trigger modified_at (fungsi update_modified_at_column sudah ada dari migration 002).
DO $$
DECLARE t text;
BEGIN
  FOREACH t IN ARRAY ARRAY['m_product', 'm_product_variant'] LOOP
    EXECUTE format('DROP TRIGGER IF EXISTS update_%1$s_modified_at ON %1$s;', t);
    EXECUTE format('CREATE TRIGGER update_%1$s_modified_at BEFORE UPDATE ON %1$s FOR EACH ROW EXECUTE FUNCTION update_modified_at_column();', t);
  END LOOP;
END $$;
