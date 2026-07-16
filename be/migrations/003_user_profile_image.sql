-- Migration 003 — tambah kolom profile_image (URL foto profil) ke m_internal_user.
-- Menyimpan link gambar (opsional).

ALTER TABLE m_internal_user ADD COLUMN IF NOT EXISTS profile_image text;
