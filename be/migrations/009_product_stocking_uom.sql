-- Kolom stocking_uom: UoM yang dipakai untuk pelaporan/penghitungan stok
-- (ditampilkan di inventory report). Referensi ke m_uom, boleh kosong.
ALTER TABLE m_product ADD COLUMN IF NOT EXISTS stocking_uom uuid REFERENCES m_uom(id);
