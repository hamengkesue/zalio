-- Migration 005 — Product master tables: brand, category, subcategory, uom.
-- Mengikuti pola m_internal_user (uuid, kolom audit, trigger modified_at).

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

-- Trigger modified_at untuk tiap tabel
-- (fungsi update_modified_at_column sudah dibuat di migration 002).
DO $$
DECLARE t text;
BEGIN
  FOREACH t IN ARRAY ARRAY['m_brand','m_category','m_subcategory','m_uom'] LOOP
    EXECUTE format('DROP TRIGGER IF EXISTS update_%1$s_modified_at ON %1$s;', t);
    EXECUTE format('CREATE TRIGGER update_%1$s_modified_at BEFORE UPDATE ON %1$s FOR EACH ROW EXECUTE FUNCTION update_modified_at_column();', t);
  END LOOP;
END $$;
