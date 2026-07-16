-- Migration 004 — tambah kolom group_access (placeholder) ke m_internal_user.
-- Untuk sekarang berupa teks bebas; opsi/aturan ditentukan menyusul.

ALTER TABLE m_internal_user ADD COLUMN IF NOT EXISTS group_access text;
