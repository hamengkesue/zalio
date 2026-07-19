-- Migration 007 — cascade: kalau m_category dinonaktifkan (is_active true->false),
-- semua m_subcategory di bawahnya ikut nonaktif.
-- Mengaktifkan kembali category TIDAK mengaktifkan subcategory otomatis.

CREATE OR REPLACE FUNCTION cascade_category_inactive()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
  IF NEW.is_active = false AND OLD.is_active = true THEN
    UPDATE m_subcategory
    SET is_active = false
    WHERE category_id = NEW.id AND is_active = true;
  END IF;
  RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_category_cascade_inactive ON m_category;
CREATE TRIGGER trg_category_cascade_inactive
  AFTER UPDATE OF is_active ON m_category
  FOR EACH ROW EXECUTE FUNCTION cascade_category_inactive();
