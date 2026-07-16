-- Migration 002 — Internal users (m_internal_user)
-- Sumber kebenaran untuk tabel pengguna internal back-office.
-- id UUID, kolom audit (created_by/modified_by), trigger modified_at,
-- role admin/staff, username & email unik.

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Trigger function: set modified_at = now() setiap UPDATE.
CREATE OR REPLACE FUNCTION update_modified_at_column()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    NEW.modified_at = now();
    RETURN NEW;
END;
$$;

CREATE TABLE IF NOT EXISTS m_internal_user (
    id              uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    full_name       text NOT NULL,
    email           text NOT NULL UNIQUE,
    whatsapp_number text,
    username        text NOT NULL,
    password_hash   text NOT NULL,
    role            text NOT NULL DEFAULT 'staff' CHECK (role IN ('admin', 'staff')),
    is_active       boolean NOT NULL DEFAULT true,
    created_at      timestamptz NOT NULL DEFAULT now(),
    created_by      uuid REFERENCES m_internal_user(id) ON DELETE SET NULL,
    modified_at     timestamptz NOT NULL DEFAULT now(),
    modified_by     uuid REFERENCES m_internal_user(id) ON DELETE SET NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_internal_user_username ON m_internal_user (username);

DROP TRIGGER IF EXISTS update_m_internal_user_modified_at ON m_internal_user;
CREATE TRIGGER update_m_internal_user_modified_at
    BEFORE UPDATE ON m_internal_user
    FOR EACH ROW EXECUTE FUNCTION update_modified_at_column();

-- Admin default dibuat otomatis oleh backend saat pertama jalan
-- (lihat internal/modules/auth/seed.go).
