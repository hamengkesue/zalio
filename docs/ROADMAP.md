# Roadmap: ERP Back-Office + Ecommerce (Retail/Commerce & Fulfillment)

> Dokumen perencanaan untuk membangun proyek ERP baru (`zalio_erp`) dari 0 lewat prompting di Claude Code, oleh product owner non-coder.
> Tanggal: 2026-07-15

---

## 1. Keputusan Inti

| Aspek | Keputusan | Alasan |
|-------|-----------|--------|
| **Arsitektur** | Monolith modular | Cepat jalan, mudah di-debug, tetap bisa dipecah jadi microservice nanti |
| **Backend** | Go + Gin + pgx | Sama seperti `fulka-ffc`, ada contoh kerja untuk ditiru |
| **Frontend** | Nuxt 4 + `@nuxt/ui` (Vue) | Sama seperti `fulka-ffc` |
| **Database** | PostgreSQL | Sama seperti `fulka-ffc` |
| **File storage** | MinIO (opsional, saat butuh upload) | Sama seperti `fulka-ffc` |
| **Infra** | Docker Compose (lokal) | Sama seperti `fulka-ffc` |
| **Domain** | Retail/commerce & fulfillment | Pinjam pola & skema dari `fulka-ffc` |
| **Fokus awal** | ERP back-office untuk admin | Storefront ecommerce menyusul (Fase 9) |
| **Menuju microservice** | Ya, tapi via monolith modular dulu | Ekstraksi modul saat benar-benar perlu skala |

---

## 2. Dua Prinsip yang Menyetir Semuanya

### Prinsip A — Vertical Slice
Tiap fitur dibangun **tembus dari database sampai tampilan**, dalam potongan kecil yang bisa dilihat jalan:

```
migration (tabel) → model (struct) → repository (baca/tulis DB)
   → handler (API) → route (daftar di main.go)
   → composable (frontend ambil data) → page (halaman)
```

Bukan "bikin semua tabel dulu, semua API dulu". Tiap slice = satu hal yang bisa kamu buka di browser dan buktikan bekerja.

### Prinsip B — Modular (siap jadi microservice)
Tiap modul (Produk, Inventory, Purchasing, ...) punya **folder & batas sendiri**, komunikasi antar-modul lewat antarmuka yang jelas — bukan saling sikut isi dalamnya. Disiplin ini yang bikin nanti gampang mencabut satu modul jadi service terpisah.

**Struktur modular yang diusulkan (backend):**
```
be/internal/
├── modules/
│   ├── auth/         { model, repository, handler, routes }
│   ├── product/      { model, repository, handler, routes }
│   ├── inventory/    { ... }
│   ├── purchasing/   { ... }
│   ├── sales/        { ... }
│   └── ...
├── platform/         # hal lintas-modul: db pool, config, minio, middleware
└── cmd/api/main.go   # rakit semua modul
```

> Catatan: `fulka-ffc` menaruh semua repository/handler di satu folder datar. Untuk proyek baru yang menuju microservice, pisahkan per-`modules/` sejak awal.

---

## 3. Aturan Prompting (berlaku di semua fase)

Salin kebiasaan ini ke setiap sesi:

1. **Satu prompt = satu hasil yang bisa dilihat.** Berhenti tiap langkah untuk dicek.
2. **Selalu minta bukti.** "Jalankan dan tunjukkan bahwa X bekerja (screenshot/log/test)."
3. **Minta penjelasan awam.** "Jelaskan kenapa, anggap aku tidak bisa ngoding."
4. **Jaga konsistensi antar-sesi.** "Ikuti pola modul yang sudah ada, jangan bikin gaya baru."
5. **Simpan keputusan penting** ke memory/dokumen supaya sesi berikutnya konsisten.

**Prompt pembuka tiap fitur (template):**
> "Kita bangun [FITUR] di modul [MODUL]. Ikuti pola vertical-slice dan struktur `modules/` yang sudah ada. Kerjakan bertahap: (1) migration, (2) model, (3) repository, (4) handler + route, (5) composable + page. Berhenti setiap langkah biar aku cek. Di akhir, jalankan dan tunjukkan bukti aku bisa [AKSI] lewat browser. Jelaskan tiap langkah dalam bahasa awam."

---

## 4. Fase-Fase Roadmap

Urutan mengikuti aturan **"yang jadi rujukan dibangun duluan"**.

### Fase 0 — Fondasi & Scaffold
**Tujuan:** menyiapkan kerangka proyek + membuktikan pola vertical-slice jalan lewat 1 contoh sederhana.
**Hasil yang bisa dilihat:** aplikasi kosong yang bisa dibuka di browser, punya health-check API, dan 1 halaman contoh ("Hello data") yang mengambil data dari database.

**Contoh prompt:**
> "Buatkan proyek `zalio_erp`: monolith modular dengan Go+Gin+pgx (backend) dan Nuxt 4 + @nuxt/ui (frontend), PostgreSQL + docker-compose, meniru struktur `fulka-ffc` TAPI backend pakai struktur `internal/modules/` + `internal/platform/`. Sediakan: docker-compose (Postgres saja dulu), backend dengan endpoint `/api/v1/health`, frontend dengan 1 halaman dashboard kosong + layout sidebar. Jalankan semuanya dan tunjukkan health-check hijau + halaman terbuka di browser. Jelaskan tiap file yang kamu buat."

Lalu, satu slice contoh:
> "Tambahkan modul contoh `ping`: tabel `tb_ping` berisi pesan, endpoint list-nya, dan halaman yang menampilkannya. Ini untuk membuktikan pola vertical-slice. Setelah jalan, jelaskan pola ini supaya aku paham cara modul berikutnya dibuat."

---

### Fase 1 — Auth, User & Role (RBAC)
**Tujuan:** back-office wajib punya login + hak akses. (Ini yang `fulka-ffc` belum punya.)
**Hasil yang bisa dilihat:** halaman login; setelah login, admin masuk dashboard; user dengan role berbeda melihat menu berbeda.

**Konsep yang perlu kamu paham:** *autentikasi* (siapa kamu — login) vs *otorisasi* (kamu boleh apa — role/permission). Password disimpan ter-*hash*, sesi pakai token (JWT).

**Contoh prompt:**
> "Bangun modul `auth`: registrasi user internal (oleh admin), login dengan email+password (password di-hash pakai bcrypt), dan token JWT untuk sesi. Tambahkan tabel `user`, `role`, `permission`, dan relasinya. Buat middleware yang memblokir endpoint kalau belum login. Sediakan halaman login di frontend + simpan token, dan proteksi semua halaman lain. Jangan pernah menaruh password dalam bentuk teks biasa. Setelah jalan, tunjukkan: login berhasil → dashboard, akses tanpa login → ditolak."

> "Tambahkan role sederhana: `admin` dan `staff`. Menu sidebar tampil sesuai role. Tunjukkan dengan 2 user berbeda bahwa menunya beda."

> ⚠️ **Batasan penting:** Claude Code tidak boleh mengetik/menyimpan password atau kredensial asli untukmu. Untuk data uji, pakai password dummy; untuk produksi nanti, kamu sendiri yang mengaturnya.

---

### Fase 2 — Master Data
**Tujuan:** membangun "kata benda" yang dirujuk semua modul lain.
**Sub-modul (urutan disarankan):** Kategori → Produk (+ Varian) → Supplier/Vendor → Gudang (+ Lokasi) → Customer.
**Hasil yang bisa dilihat:** tiap entitas punya halaman list + tambah/edit/nonaktifkan.

**Contoh prompt (ulangi pola ini per entitas):**
> "Bangun modul `product` (vertical slice penuh): tabel produk dengan nama, SKU, kategori, harga, satuan, status aktif; plus tabel varian produk. CRUD lengkap (list dengan pencarian & pagination, tambah, edit, toggle-aktif). Frontend: halaman daftar produk + form tambah/edit pakai komponen @nuxt/ui. Ikuti pola modul sebelumnya. Tunjukkan aku bisa menambah produk baru dan melihatnya di daftar."

> "Sekarang ulangi pola yang sama untuk modul `supplier` (nama, kontak, alamat, status) dan `warehouse` (nama, kode, lokasi). Jaga konsistensi dengan modul product."

---

### Fase 3 — Inventory / Stok
**Tujuan:** melacak stok per produk per gudang.
**Hasil yang bisa dilihat:** halaman stok menampilkan jumlah per produk/gudang; bisa lakukan penyesuaian stok (stock adjustment) dan lihat riwayatnya.

**Konsep:** stok berubah karena *transaksi* (barang masuk/keluar), bukan diedit langsung. Simpan tiap pergerakan sebagai baris (batch/ledger) supaya bisa ditelusuri.

**Contoh prompt:**
> "Bangun modul `inventory`: tabel `inventory_batch` (produk, gudang, jumlah, harga modal, tanggal) dan `stock_movement` (jenis: masuk/keluar/adjust, jumlah, referensi, waktu). Buat fungsi hitung stok tersedia per produk per gudang. Frontend: halaman ringkasan stok + form penyesuaian stok manual dengan alasan. Tunjukkan: setelah adjust +10, stok bertambah dan muncul di riwayat pergerakan."

---

### Fase 4 — Purchasing (Procurement)
**Tujuan:** proses beli barang ke supplier, dan barang yang diterima **menambah stok**.
**Alur:** Purchase Request (PR) → Purchase Order (PO) → Goods Receipt (terima barang → stok naik).
**Hasil yang bisa dilihat:** buat PR, ubah jadi PO ke supplier, terima barang, dan lihat stok inventory bertambah otomatis.

**Contoh prompt:**
> "Bangun modul `purchasing`: (1) Purchase Request — daftar produk yang mau dibeli; (2) Purchase Order — dibuat dari PR, ditujukan ke supplier, dengan harga & jumlah; (3) Goods Receipt — saat barang datang, catat penerimaan yang OTOMATIS membuat inventory_batch baru (stok naik). Pakai transaksi DB supaya konsisten. Frontend: halaman PR, PO, dan penerimaan. Tunjukkan alur penuh: PR → PO → terima → cek stok di modul inventory sudah bertambah."

---

### Fase 5 — Sales Order
**Tujuan:** mencatat pesanan penjualan; pesanan **mengurangi/booking stok**.
**Hasil yang bisa dilihat:** buat sales order, pilih produk & jumlah, sistem cek stok, total terhitung otomatis.

**Contoh prompt:**
> "Bangun modul `sales`: Sales Order dengan customer, daftar produk + jumlah + harga, total otomatis (pakai transaksi DB seperti pola `fulfillment_orders` di fulka). Saat order dibuat, tandai stok ter-booking. Frontend: halaman daftar order + form buat order dengan pemilihan produk. Tunjukkan: buat order 3 item → total benar → status stok berubah."

---

### Fase 6 — Fulfillment (Picking, Packing, Shipping)
**Tujuan:** memproses order jadi barang terkirim (inti fulka).
**Alur:** Order dibayar/dikonfirmasi → Picking (ambil dari gudang, stok keluar) → Packing → Shipping (input resi/kurir).
**Hasil yang bisa dilihat:** order berjalan melewati tahap-tahap; stok berkurang saat picking; status kirim terlacak.

**Contoh prompt:**
> "Bangun modul `fulfillment`: alur status order Picking → Packing → Shipped. Saat picking, kurangi stok inventory (buat stock_movement keluar). Simpan alamat kirim & info kurir/resi. Frontend: papan proses fulfillment per order dengan tombol pindah tahap. Tunjukkan: order lewat semua tahap, stok berkurang saat picking, resi tersimpan."

---

### Fase 7 — Finance Dasar
**Tujuan:** uang mengikuti barang — invoice & pembayaran (versi sederhana, bukan akuntansi penuh).
**Hasil yang bisa dilihat:** buat invoice dari sales order, catat pembayaran, lihat status lunas/belum.

**Contoh prompt:**
> "Bangun modul `finance` sederhana: Invoice dibuat dari Sales Order (nomor, jatuh tempo, total), dan pencatatan Payment (jumlah, tanggal, metode). Status invoice: unpaid/partial/paid dihitung dari total pembayaran. Frontend: daftar invoice + form catat pembayaran. Tunjukkan: invoice unpaid → catat pembayaran penuh → status jadi paid."

> ⚠️ Ini bukan sistem akuntansi resmi. Kalau butuh pembukuan/pajak sungguhan, konsultasikan dengan akuntan.

---

### Fase 8 — Dashboard & Laporan
**Tujuan:** ringkasan dari semua modul untuk pengambilan keputusan.
**Hasil yang bisa dilihat:** dashboard dengan kartu angka (penjualan hari ini, stok menipis, PO menunggu) + beberapa laporan/tabel yang bisa difilter.

**Contoh prompt:**
> "Bangun dashboard: kartu ringkasan (total penjualan bulan ini, jumlah order per status, produk stok menipis, PO belum diterima). Tambahkan halaman laporan penjualan yang bisa difilter per tanggal & produk. Ambil data dari modul-modul yang sudah ada tanpa mengubah isinya. Tunjukkan angka di dashboard cocok dengan data di modul masing-masing."

---

### Fase 9 — Storefront Ecommerce (customer-facing)
**Tujuan:** website tempat customer belanja; pesanan mereka **masuk sebagai Sales Order** di back-office.
**Hasil yang bisa dilihat:** situs publik menampilkan katalog produk, keranjang, checkout → membuat sales order yang muncul di admin.

**Konsep:** storefront ini **frontend terpisah** yang memakai backend yang sama (atau sebagian API-nya). Ini titik alami pertama di mana pemisahan modul mulai terasa berguna.

**Contoh prompt:**
> "Bangun storefront ecommerce sebagai frontend Nuxt terpisah yang memakai API produk & order dari backend yang sama. Fitur: halaman katalog (hanya produk aktif & ada stok), detail produk, keranjang, dan checkout tamu yang membuat Sales Order di backend. Jangan tampilkan data internal (harga modal, stok pemasok). Tunjukkan: customer checkout → order muncul di halaman Sales Order admin."

> ⚠️ **Pembayaran online** (payment gateway) melibatkan uang & kredensial rahasia — itu harus kamu integrasikan sendiri dengan akun merchant-mu; Claude bisa bantu kodenya tapi tidak memasukkan kunci/uang untukmu.

---

### Fase 10 — (Nanti) Ekstraksi ke Microservice
**Tujuan:** cabut **satu** modul yang benar-benar butuh skala sendiri jadi service terpisah — bukan semua sekaligus.
**Kandidat pertama biasanya:** modul dengan beban paling berat (mis. katalog/produk untuk storefront, atau inventory).
**Prasyarat:** modul itu sudah rapi batasnya (hasil disiplin Prinsip B).

**Contoh prompt:**
> "Modul `product` sekarang jadi beban terberat karena dipakai storefront. Bantu aku pisahkan modul ini jadi service terpisah: pindahkan kode ke service sendiri dengan API-nya, tentukan bagaimana modul lain memanggilnya lewat jaringan, dan jelaskan trade-off yang muncul (latency, konsistensi data). Kerjakan bertahap dan jelaskan tiap risiko dalam bahasa awam."

---

## 5. Cara Menjalankan Roadmap Ini

- **Selesaikan fase secara berurutan.** Jangan loncat; tiap fase memakai hasil fase sebelumnya.
- **Dalam satu fase, kerjakan per vertical-slice.** Satu entitas/alur, buktikan jalan, baru lanjut.
- **Di akhir tiap fase**, minta Claude merangkum apa yang sudah dibangun & memperbarui dokumentasi proyek.
- **Simpan keputusan arsitektur** (mis. "kita pakai JWT", "struktur modules/") agar sesi berikutnya konsisten.

## 6. Hal-Hal yang Kamu (Bukan Claude) Harus Tangani Sendiri
- Membuat akun & memasukkan password/kredensial asli (payment gateway, email, domain).
- Membeli domain, hosting, atau layanan berbayar.
- Keputusan legal/keuangan (pajak, pembukuan resmi).
- Menyetujui hal yang tak bisa dibatalkan (deploy ke publik, hapus data).

## 7. Referensi
- Repo contoh pola: `fulka-ffc` (README-nya menjelaskan pola vertical-slice & cara menambah entitas).
- Pola transaksi DB dengan total terhitung: lihat `fulfillment_orders` di `fulka-ffc`.
