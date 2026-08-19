# Zalio ERP — Setup Ulang di Device Baru

Panduan langkah-demi-langkah untuk menjalankan kembali project ini di komputer baru,
lengkap dengan cara mengembalikan **database** dan **file gambar (MinIO)** dari backup.

---

## Gambaran Singkat (apa jalan di mana)

| Bagian | Teknologi | Cara jalan | Alamat |
|---|---|---|---|
| Database | PostgreSQL 16 (Docker) | `docker compose up` | localhost:**5435** |
| Penyimpanan file | MinIO (Docker) | `docker compose up` | API localhost:**9004**, Console localhost:**9005** |
| Backend | Go 1.26 | `go run ./cmd/api` | localhost:**8082** |
| Frontend | Nuxt + Yarn | `yarn dev` | localhost:**3000** |

Backend & frontend **jalan dari source code** (tidak di Docker). Docker hanya untuk DB + MinIO.

---

## Yang Perlu Disiapkan

**Dari Google Drive (folder backup `zalio-backup`) — sumber utama, PALING LENGKAP:**
- `zalio-erp-source.zip` — **seluruh kode** (termasuk perubahan yang belum di-commit ke GitHub) + `.env` + riwayat git
- `zalio_db.sql` — isi database (produk, brand, COA, dll)
- `minio_data.tar.gz` — semua file gambar

> ⚠️ **Penting:** ada kode yang belum di-commit ke GitHub. Jadi di device baru, **pakai `zalio-erp-source.zip`**,
> jangan `git clone` (clone hanya berisi kode yang sudah di-commit, akan ketinggalan pekerjaan terakhir).

---

## STEP 0 — Install Prasyarat (sekali saja)

👉 **Panduan install lengkap & detail ada di `INSTALL.md`** (Docker Desktop, DBeaver, Claude Code, Go, Node, Yarn, Git).

Ringkasnya, yang WAJIB ada sebelum lanjut:
1. **Docker Desktop** — https://www.docker.com/products/docker-desktop → buka sampai statusnya "running".
2. **Go** (1.26+) — https://go.dev/dl/ → cek `go version`.
3. **Node.js** (20 LTS+) — https://nodejs.org → cek `node -v`.
4. **Yarn** — `npm install -g yarn` → cek `yarn -v`.
5. **Git** — `git --version` (biasanya sudah ada).

Opsional: **DBeaver** (lihat isi database) & **Claude Code** (lanjut develop dengan AI) — cara install di `INSTALL.md`.

---

## STEP 1 — Ambil Kode (unzip dari Drive)

```bash
# masuk ke folder tempat menaruh project, contoh:
mkdir -p ~/myProject && cd ~/myProject
# unzip file dari Drive (misal ada di ~/Downloads). Ini akan membuat folder 'zalio_erp'.
unzip ~/Downloads/zalio-erp-source.zip
cd zalio_erp
```

Karena zip ini sudah termasuk `.env` dan riwayat git, folder langsung siap pakai (skip STEP 2 kalau `.env` sudah ada — cek `cat be/.env`).

> ⚠️ **Nama folder harus `zalio_erp`** (zip sudah otomatis begitu). Nama volume Docker mengikuti nama folder ini
> (mis. `zalio_erp_minio_data`). Kalau berbeda, langkah restore MinIO perlu disesuaikan.
>
> *Alternatif:* `git clone <URL-REPO>` — TAPI hanya berisi kode yang sudah di-commit (bisa ketinggalan pekerjaan terakhir). Gunakan zip untuk yang paling lengkap.

---

## STEP 2 — Kembalikan Secret Config (.env)

Copy `be.env.txt` dari Drive ke lokasi yang benar dengan nama `.env`:

```bash
# misal file backup ada di ~/Downloads
cp ~/Downloads/be.env.txt be/.env
```

Cek isinya benar: `cat be/.env` — harus ada `DB_PORT=5435`, `MINIO_ENDPOINT=localhost:9004`, dll.

---

## STEP 3 — Nyalakan Database + MinIO (Docker)

```bash
# dari dalam folder zalio_erp (yang ada docker-compose.yml)
docker compose up -d
# tunggu ~10 detik biar PostgreSQL siap menerima koneksi
```

Cek kedua container jalan:
```bash
docker ps
# harus muncul: zalio-erp-db dan zalio-erp-minio
```

Saat ini database & MinIO masih **kosong** — kita isi dari backup di step berikut.

---

## STEP 4 — Restore Database

```bash
docker exec -i zalio-erp-db psql -U zalio_erp -d zalio_erp < ~/Downloads/zalio_db.sql
```

> ℹ️ File `zalio_db.sql` sudah berisi **struktur tabel + data**, jadi **JANGAN jalankan migration**
> secara terpisah. Cukup restore file ini ke database yang masih kosong.

Cek berhasil (harus keluar angka jumlah produk):
```bash
docker exec -it zalio-erp-db psql -U zalio_erp -d zalio_erp -c "SELECT count(*) FROM m_product;"
```

---

## STEP 5 — Restore File MinIO (gambar)

```bash
docker compose stop minio
docker run --rm -v zalio_erp_minio_data:/data -v ~/Downloads:/backup alpine sh -c "cd /data && tar xzf /backup/minio_data.tar.gz"
docker compose start minio
```

> Kalau nama folder project bukan `zalio_erp`, ganti `zalio_erp_minio_data` dengan nama yang benar.
> Cek nama volume dengan: `docker volume ls | grep minio`.

---

## STEP 6 — Jalankan Backend

```bash
cd be
go mod download        # unduh dependency (sekali saja / saat ada perubahan)
go run ./cmd/api
```

Kalau sukses, akan muncul `Server starting on :8082`. **Biarkan terminal ini tetap terbuka.**

---

## STEP 7 — Jalankan Frontend

Buka **terminal baru**:

```bash
cd ~/myProject/zalio_erp/fe
yarn install           # unduh dependency (sekali saja / saat ada perubahan)
yarn dev
```

Akan muncul alamat seperti `http://localhost:3000`.

---

## STEP 8 — Buka & Cek

1. Buka browser ke **http://localhost:3000**.
2. Login pakai akun Anda.
3. Cek menu **Product Management** — data produk + gambar harus muncul (tanda DB & MinIO ter-restore).
4. (Opsional) Console MinIO: http://localhost:9005 (user/pass ada di `be/.env`).

Selesai! 🎉

---

## Menjalankan Sehari-hari (setelah setup awal)

**Menyalakan:**
```bash
cd ~/myProject/zalio_erp
docker compose up -d              # DB + MinIO
cd be && go run ./cmd/api         # backend (terminal 1)
cd fe && yarn dev                 # frontend (terminal 2)
```

**Mematikan:**
- Backend & frontend: tekan `Ctrl + C` di masing-masing terminal.
- DB + MinIO: `docker compose stop` (data tetap aman) — atau biarkan jalan.

---

## Cara Membuat Backup Baru (rutin, biar aman)

Cukup jalankan **skrip 1-klik** dari folder `zalio_erp` (pastikan `docker compose up -d` sudah jalan):
```bash
./backup.sh
```
Skrip otomatis membuat ke `~/zalio-backup/`: dump database, arsip file MinIO, dan zip kode (tanpa node_modules) + `.env`.
Lalu tinggal **drag folder `~/zalio-backup` ke Google Drive**. Lakukan berkala (mis. tiap selesai kerja besar).

> 💡 Pakai Claude Code? Cukup bilang: **"jalankan backup.sh"** atau **"backup project zalio"**.

---

## Setup TANPA Backup (mulai dari nol / database kosong)

Kalau ingin mulai bersih tanpa data lama (skip STEP 4 & 5), jalankan migration + seed:
```bash
# setelah 'docker compose up -d', apply semua migration berurutan:
for f in be/migrations/*.sql; do docker exec -i zalio-erp-db psql -U zalio_erp -d zalio_erp < "$f"; done
```
Bucket MinIO otomatis dibuat backend saat pertama jalan. Data master (brand/COA/dll) harus diisi manual/seed.

---

## Troubleshooting

| Masalah | Solusi |
|---|---|
| `docker: command not found` | Docker Desktop belum jalan / belum diinstall. Buka aplikasinya. |
| Backend error "connection refused" ke DB | Tunggu Postgres siap (~10 detik) atau cek `docker ps`. Pastikan `be/.env` benar (`DB_PORT=5435`). |
| Frontend jalan tapi data kosong / gambar tidak muncul | Backend belum jalan, atau DB/MinIO belum di-restore. |
| Port bentrok (5435/9004/8082/3000 dipakai) | Matikan aplikasi lain yang pakai port itu, atau ubah port di `docker-compose.yml` / `be/.env`. |
| `psql` saat restore banyak "already exists" | Database tidak kosong. Reset dulu: `docker compose down -v && docker compose up -d`, lalu restore ulang. |
| Nama volume MinIO tidak ketemu | `docker volume ls` → pakai nama yang berakhiran `_minio_data`. |

---

*Dokumen ini juga tersimpan di folder backup `~/zalio-backup` bersama file DB & MinIO.*
