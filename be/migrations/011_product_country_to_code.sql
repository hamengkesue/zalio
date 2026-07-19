-- country_of_origin sekarang menyimpan kode ISO (mis. "ID"), bukan nama.
-- Konversi data lama yang masih berupa nama negara menjadi kode-nya.
UPDATE m_product p
SET country_of_origin = c.country_code
FROM m_country c
WHERE p.country_of_origin IS NOT NULL
  AND lower(p.country_of_origin) = lower(c.country_name);
