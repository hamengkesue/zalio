-- Migration 002 — Autentikasi & pengguna (Fase 1)
-- Tabel user internal untuk back-office. Password disimpan sebagai HASH (bcrypt),
-- tidak pernah dalam bentuk teks biasa. Role sederhana: admin / staff.

CREATE TABLE IF NOT EXISTS tb_user (
    id            SERIAL PRIMARY KEY,
    name          TEXT NOT NULL,
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role          TEXT NOT NULL DEFAULT 'staff' CHECK (role IN ('admin', 'staff')),
    is_active     BOOLEAN NOT NULL DEFAULT true,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    modified_at   TIMESTAMPTZ
);

-- Admin default dibuat otomatis oleh backend saat pertama kali jalan
-- (lihat internal/modules/auth/seed.go), jadi tidak di-seed di sini.
