# Panduan Akses Database & MinIO — Zalio ERP

Cara masuk ke database (lewat DBeaver) dan penyimpanan file (MinIO) untuk proyek `zalio_erp`.

> Semua kredensial di bawah **hanya untuk lokal/development**. Ganti sebelum dipakai sungguhan.
> Pastikan layanannya hidup dulu: jalankan `docker compose up -d` di folder proyek.

---

## 🗄️ Masuk ke Database (DBeaver)

Di DBeaver: **Database → New Connection → PostgreSQL**, isi:

| Field | Nilai |
|---|---|
| Host | `localhost` |
| Port | `5435` |
| Database | `zalio_erp` |
| Username | `zalio_erp` |
| Password | `zalio_erp_secret` |

Klik **Test Connection → Finish**.

Tabel ada di schema **`public`**: `tb_ping`, `tb_user`.

> Kalau DBeaver minta download driver PostgreSQL, klik **Download** saja (sekali).

---

## 📦 Masuk ke MinIO (penyimpanan gambar/video)

Buka **console** di browser: **http://localhost:9005**

| Field | Nilai |
|---|---|
| Username | `zalio_erp_minio` |
| Password | `zalio_erp_minio_secret` |

Bucket **`zalio-erp`** sudah dibuat otomatis — langsung kelihatan di menu **Buckets / Object Browser**, dan kamu bisa upload file di situ.

**Detail teknis MinIO** (untuk nanti dipakai aplikasi):

| Item | Nilai |
|---|---|
| S3 API endpoint | `localhost:9004` |
| Console | `localhost:9005` |
| Bucket | `zalio-erp` |

---

## Menyalakan / mematikan layanan

```bash
# di folder proyek: /Users/user/myProject/zalio_erp_project/zalio_erp

docker compose up -d      # nyalakan database + MinIO
docker compose ps         # lihat status
docker compose stop       # matikan (data tetap tersimpan)
```
