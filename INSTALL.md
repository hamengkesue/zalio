# Install Tools di Mac Baru — Zalio ERP

Panduan install semua alat yang dibutuhkan. Setelah selesai, lanjut ke **`SETUP.md`** untuk
mengembalikan data & menjalankan project.

## Ringkasan

| Tool | Wajib? | Fungsi |
|---|---|---|
| **Docker Desktop** | ✅ WAJIB | Menjalankan database (PostgreSQL) + MinIO |
| **DBeaver** | ⭐ Opsional (disarankan) | Melihat / mengedit isi database secara visual |
| **Claude Code** | ⭐ Disarankan | Lanjut develop dibantu AI |
| **Go, Node + Yarn, Git** | ✅ WAJIB | Menjalankan backend & frontend |

---

## 1. Docker Desktop (WAJIB)

Database + MinIO jalan di dalam Docker, jadi ini wajib.

1. Buka **https://www.docker.com/products/docker-desktop**
2. Klik **Download for Mac** — pilih chip yang sesuai:
   - Mac baru (M1/M2/M3/M4) → **Apple Silicon**
   - Mac lama (Intel) → **Intel Chip**
   > Cara cek: menu  → About This Mac → lihat "Chip".
3. Buka file `.dmg` yang terunduh → drag **Docker** ke folder **Applications**.
4. Buka aplikasi **Docker** dari Applications → tunggu sampai ikon paus di menu bar **hijau/stabil** ("Docker Desktop is running").
5. Setujui permission yang diminta (sekali di awal).

**Cek berhasil** — buka Terminal, ketik:
```bash
docker --version
docker compose version
```
Kalau keluar nomor versi (bukan "command not found"), berhasil. ✅

> 💡 Docker Desktop **gratis** untuk pemakaian pribadi / perusahaan kecil.

---

## 2. DBeaver (opsional, tapi enak buat lihat database)

Aplikasi gratis untuk melihat/mengedit isi database lewat tampilan tabel (tanpa perlu ketik SQL).

**Install:**
1. Buka **https://dbeaver.io/download/** → pilih **DBeaver Community Edition** → **macOS (dmg)** (pilih Apple Silicon / Intel sesuai chip).
2. Buka `.dmg` → drag **DBeaver** ke **Applications** → buka aplikasinya.

**Hubungkan ke database Zalio** (jalankan setelah `docker compose up -d` aktif):
1. Menu **Database → New Database Connection** → pilih **PostgreSQL** → Next.
2. Isi:
   | Kolom | Nilai |
   |---|---|
   | Host | `localhost` |
   | Port | `5435` |
   | Database | `zalio_erp` |
   | Username | `zalio_erp` |
   | Password | `zalio_erp_secret` |
3. Centang **Save password**.
4. Klik **Test Connection** → kalau diminta download driver PostgreSQL, klik **Download**.
5. Kalau "Connected", klik **Finish**.

Sekarang di panel kiri bisa buka: `zalio_erp → Schemas → public → Tables` → klik kanan sebuah tabel → **View Data** untuk lihat isinya (mis. `m_product`, `m_brand`, `m_coa`).

> ⚠️ Password di atas (`zalio_erp_secret`) berasal dari `docker-compose.yml`. Kalau nanti diganti, sesuaikan.

---

## 3. Claude Code (untuk lanjut develop dengan AI)

**Prasyarat:** Node.js sudah terinstall (lihat bagian 4). Butuh langganan Claude (Pro/Max) atau API key Anthropic.

**Install (pilih salah satu):**

- **Cara A — installer resmi (paling gampang):**
  ```bash
  curl -fsSL https://claude.com/install.sh | bash
  ```
- **Cara B — lewat npm (kalau sudah ada Node):**
  ```bash
  npm install -g @anthropic-ai/claude-code
  ```

**Cek berhasil:**
```bash
claude --version
```

**Pakai pertama kali:**
1. Masuk ke folder project:
   ```bash
   cd ~/myProject/zalio_erp
   ```
2. Jalankan:
   ```bash
   claude
   ```
3. Ikuti proses login yang muncul (login lewat browser pakai akun Claude Anda, atau masukkan API key).
4. Selesai — sekarang bisa mulai minta bantuan, mis. ketik: `backup project zalio` atau `jalankan project`.

> Dokumentasi resmi (kalau ada perubahan): **https://docs.claude.com/claude-code**

---

## 4. Tools Dev Lain (WAJIB — untuk menjalankan backend & frontend)

| Tool | Download | Cek |
|---|---|---|
| **Git** | biasanya sudah ada; kalau belum: https://git-scm.com | `git --version` |
| **Go** (1.26+) | https://go.dev/dl/ (pilih macOS pkg sesuai chip) | `go version` |
| **Node.js** (20 LTS+) | https://nodejs.org (pilih **LTS**) | `node -v` |
| **Yarn** | setelah Node: `npm install -g yarn` | `yarn -v` |

Semua installer di atas berbentuk `.pkg`/`.dmg` — tinggal buka & Next sampai selesai.

---

## Sudah Semua? Lanjut ke `SETUP.md`

Setelah semua tool ✅, buka **`SETUP.md`** untuk: unzip kode → restore database → restore file MinIO → jalankan backend & frontend.
