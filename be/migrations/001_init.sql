-- Migration 001 — skema awal zalio_erp
-- Fase 0: tabel contoh `tb_ping` untuk membuktikan pola vertical-slice
-- (database -> repository -> handler -> API -> halaman).

CREATE TABLE IF NOT EXISTS tb_ping (
    id         SERIAL PRIMARY KEY,
    message    TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO tb_ping (message) VALUES
    ('Halo dari zalio_erp — backend & database tersambung!'),
    ('Ini baris contoh kedua dari database.');
