-- Gambar utama per-varian (default = main image induk produk).
-- Ditampilkan & diedit di kolom Image pada Variant List.
ALTER TABLE m_product_variant ADD COLUMN IF NOT EXISTS variant_image text;
