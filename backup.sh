#!/usr/bin/env bash
#
# Zalio ERP — Backup 1-klik
# ─────────────────────────
# Menyimpan SEMUA yang dibutuhkan untuk pindah/develop di device lain:
#   • kode lengkap (termasuk yang belum di-commit) + .env   → zalio-erp-source.zip
#   • isi database                                          → zalio_db.sql
#   • file MinIO (gambar)                                   → minio_data.tar.gz
#   • panduan setup                                         → SETUP.md
# Hasil ada di folder  ~/zalio-backup/  — tinggal di-drag ke Google Drive.
#
# Cara pakai:  ./backup.sh      (pastikan Docker + container DB/MinIO sedang jalan)

set -euo pipefail

# ── Konfigurasi ──
DB_CONTAINER="zalio-erp-db"
MINIO_CONTAINER="zalio-erp-minio"
DB_USER="zalio_erp"
DB_NAME="zalio_erp"
PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"   # folder skrip = root project
OUT_DIR="$HOME/zalio-backup"

info() { printf "\033[1;34m→ %s\033[0m\n" "$1"; }
ok()   { printf "\033[1;32m✓ %s\033[0m\n" "$1"; }
die()  { printf "\033[1;31m✗ %s\033[0m\n" "$1" >&2; exit 1; }

# ── Cek prasyarat ──
command -v docker >/dev/null 2>&1 || die "Docker belum terinstall / tidak di PATH."
docker ps --format '{{.Names}}' | grep -q "^${DB_CONTAINER}$"    || die "Container '${DB_CONTAINER}' tidak jalan. Jalankan 'docker compose up -d' dulu."
docker ps --format '{{.Names}}' | grep -q "^${MINIO_CONTAINER}$" || die "Container '${MINIO_CONTAINER}' tidak jalan. Jalankan 'docker compose up -d' dulu."

STAGE="$(mktemp -d)"
trap 'rm -rf "$STAGE"' EXIT
STAMP="$(date '+%Y-%m-%d %H:%M:%S')"

# ── 1. Database ──
info "Dump database…"
docker exec "$DB_CONTAINER" pg_dump -U "$DB_USER" -d "$DB_NAME" > "$STAGE/zalio_db.sql"
ok "Database → zalio_db.sql ($(du -h "$STAGE/zalio_db.sql" | cut -f1))"

# ── 2. File MinIO (deteksi nama volume otomatis) ──
info "Arsip file MinIO…"
MINIO_VOL="$(docker inspect "$MINIO_CONTAINER" --format '{{ range .Mounts }}{{ .Name }}{{ end }}')"
[ -n "$MINIO_VOL" ] || die "Tidak menemukan volume MinIO."
docker run --rm -v "$MINIO_VOL":/data -v "$STAGE":/backup alpine tar czf /backup/minio_data.tar.gz -C /data . 2>/dev/null
ok "MinIO → minio_data.tar.gz ($(du -h "$STAGE/minio_data.tar.gz" | cut -f1))"

# ── 3. Kode (tanpa node_modules & build) ──
info "Arsip kode (tanpa node_modules)…"
( cd "$(dirname "$PROJECT_DIR")" && zip -rq "$STAGE/zalio-erp-source.zip" "$(basename "$PROJECT_DIR")" \
    -x "*/node_modules/*" "*/.nuxt/*" "*/.output/*" "*/dist/*" "*/.DS_Store" )
ok "Kode → zalio-erp-source.zip ($(du -h "$STAGE/zalio-erp-source.zip" | cut -f1))"

# ── 4. Ekstra (baca cepat tanpa unzip) ──
[ -f "$PROJECT_DIR/SETUP.md" ]           && cp "$PROJECT_DIR/SETUP.md" "$STAGE/SETUP.md"
[ -f "$PROJECT_DIR/INSTALL.md" ]         && cp "$PROJECT_DIR/INSTALL.md" "$STAGE/INSTALL.md"
[ -f "$PROJECT_DIR/docker-compose.yml" ] && cp "$PROJECT_DIR/docker-compose.yml" "$STAGE/docker-compose.yml"
[ -f "$PROJECT_DIR/be/.env" ]            && cp "$PROJECT_DIR/be/.env" "$STAGE/be.env.txt"
echo "Backup terakhir: $STAMP" > "$STAGE/LAST_BACKUP.txt"

# ── 5. Pindahkan ke folder backup (atomik: file lama tak rusak kalau gagal di tengah) ──
mkdir -p "$OUT_DIR"
cp -f "$STAGE"/* "$OUT_DIR"/

echo
ok "BACKUP SELESAI — $STAMP"
echo "Folder: $OUT_DIR"
ls -lh "$OUT_DIR"
echo
printf "\033[1;33m📤 Langkah terakhir: drag folder ~/zalio-backup ke Google Drive.\033[0m\n"
