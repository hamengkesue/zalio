# Cara Menjalankan Zalio ERP (Terminal)

Urutan menyalakan seluruh aplikasi dari nol. Butuh **3 hal jalan**: infra (database + MinIO), backend, dan frontend.

> Folder proyek: `/Users/user/myProject/zalio_erp_project/zalio_erp`
> Backend & frontend jalan di layar (foreground), jadi masing-masing butuh **jendela terminal sendiri**.

---

## Langkah 1 — Nyalakan infra (database + MinIO)

```bash
cd /Users/user/myProject/zalio_erp_project/zalio_erp
docker compose up -d
```

Cek sudah jalan:

```bash
docker compose ps
```

## Langkah 2 — Terapkan skema database (SEKALI saja, saat pertama)

> Lewati langkah ini kalau database sudah pernah di-setup.

```bash
cd /Users/user/myProject/zalio_erp_project/zalio_erp
docker exec -i zalio-erp-db psql -U zalio_erp -d zalio_erp < be/migrations/001_init.sql
docker exec -i zalio-erp-db psql -U zalio_erp -d zalio_erp < be/migrations/002_auth.sql
```

## Langkah 3 — Jalankan backend (Terminal #1)

```bash
cd /Users/user/myProject/zalio_erp_project/zalio_erp/be
go run ./cmd/api
```

Biarkan terminal ini terbuka. Backend jalan di **http://localhost:8082**.
(Pertama kali saja, kalau perlu: `go mod tidy` sebelum `go run`.)

## Langkah 4 — Jalankan frontend (Terminal #2, jendela baru)

```bash
cd /Users/user/myProject/zalio_erp_project/zalio_erp/fe
yarn dev
```

Biarkan terminal ini terbuka. Buka aplikasi di **http://localhost:3005**.
(Pertama kali saja, kalau perlu: `yarn install` sebelum `yarn dev`.)

Login: **admin@zalio.local** / **admin123**

---

## Mematikan

- **Backend / Frontend**: tekan `Ctrl + C` di masing-masing terminal.
- **Infra (database + MinIO)**:

```bash
cd /Users/user/myProject/zalio_erp_project/zalio_erp
docker compose stop      # matikan, data TETAP tersimpan
# atau
docker compose down      # matikan + hapus container (data tetap di volume)
```

---

## Ringkasan alamat

| Yang dibuka | Alamat | Login |
|---|---|---|
| Aplikasi (frontend) | http://localhost:3005 | admin@zalio.local / admin123 |
| Backend API | http://localhost:8082/api/v1/health | — |
| MinIO Console | http://localhost:9005 | zalio_erp_minio / zalio_erp_minio_secret |
| Database (DBeaver) | localhost:5435 | zalio_erp / zalio_erp_secret |
