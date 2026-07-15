-- Migration 003 — login pakai username (bukan email)
-- Menambah kolom username ke tb_user. Baris lama di-backfill dari bagian
-- depan email (mis. admin@zalio.local -> "admin").

ALTER TABLE tb_user ADD COLUMN IF NOT EXISTS username TEXT;

UPDATE tb_user
SET username = split_part(email, '@', 1)
WHERE username IS NULL OR username = '';

ALTER TABLE tb_user ALTER COLUMN username SET NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_username ON tb_user (username);
