# Zalio ERP

ERP back-office + (nanti) storefront ecommerce untuk bisnis retail/commerce & fulfillment.
Dibangun sebagai **monolith modular** yang siap dipecah jadi microservice.

- **Backend:** Go + Gin + pgx (PostgreSQL)
- **Frontend:** Nuxt 4 + `@nuxt/ui` (Nunito Sans)
- **Infra:** Docker Compose (PostgreSQL)

Peta jalan lengkap ada di [`docs/ROADMAP.md`](docs/ROADMAP.md).

| Service | URL / Port |
|---|---|
| PostgreSQL | `localhost:5435` |
| Backend API | `localhost:8082` |
| Frontend | `localhost:3005` |

> Port sengaja berbeda dari `fulka-ffc` supaya bisa jalan berdampingan.

## 1. Nyalakan infrastruktur

```bash
docker compose up -d
```

## 2. Terapkan skema database

```bash
docker exec -i zalio-erp-db psql -U zalio_erp -d zalio_erp < be/migrations/001_init.sql
```

## 3. Jalankan backend

```bash
cd be
go mod tidy   # pertama kali saja
go run ./cmd/api
```

API: http://localhost:8082 (health check: `GET /api/v1/health`).

## 4. Jalankan frontend

```bash
cd fe
yarn install  # pertama kali saja
yarn dev
```

App: http://localhost:3005.

## Struktur proyek (monolith modular)

```
zalio_erp/
├── docker-compose.yml
├── docs/ROADMAP.md            # peta jalan fase 0–10
├── be/                        # Go backend
│   ├── cmd/api/main.go        # entrypoint + rakit modul
│   ├── migrations/            # SQL migrations
│   └── internal/
│       ├── platform/          # lintas-modul: config, database
│       └── modules/           # satu folder per modul (batas jelas)
│           └── ping/          # { model, repository, handler, routes }
└── fe/                        # Nuxt frontend
    └── app/
        ├── assets/css/main.css   # design tokens
        ├── layouts/default.vue   # topbar + sidebar
        ├── components/           # AppSidebar
        ├── composables/          # usePing, useSidebar
        └── pages/                # index (dashboard), ping (contoh slice)
```

## Cara menambah modul baru (pola vertical-slice)

Tiru modul `ping`:

1. **migration** — tabel baru di `be/migrations/`
2. **model.go** — struct data
3. **repository.go** — baca/tulis DB
4. **handler.go** — handler HTTP
5. **routes.go** — fungsi `Register(rg, pool)`
6. daftarkan satu baris di `cmd/api/main.go`: `namamodul.Register(api, pool)`
7. **composable + page** di frontend

## API endpoints (Fase 0)

```
GET    /api/v1/health
GET    /api/v1/ping
POST   /api/v1/ping
```
