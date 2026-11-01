# Inventory Management REST API (Golang)

RESTful API backend untuk sistem inventaris barang dengan autentikasi JSON Web Token (JWT) dan penanganan konkurensi stok atomik, dibangun menggunakan **Go (Golang)**, framework **Gin**, **GORM**, dan database **PostgreSQL**.

Proyek ini dibuat sebagai pemenuhan Take-Home Test Backend Developer Intern.

---

## 1. Fitur Utama

- **Autentikasi & Registrasi:**
  - Registrasi pengguna baru dengan validasi ketat, normalisasi email, dan hash password aman menggunakan `bcrypt`.
  - Login mengembalikan JWT token (HS256) dengan klaim kedaluwarsa yang dapat dikonfigurasi.
- **Manajemen Inventaris (CRUD Items):**
  - Dilindungi middleware autentikasi JWT Bearer.
  - Tambah barang (`POST /items`), Detail barang (`GET /items/:id`), Ubah barang (`PUT /items/:id`), Hapus barang (`DELETE /items/:id`).
  - Pencarian fleksibel (`ILIKE` pada nama barang dengan sanitasi karakter wildcard `%` dan `_`) serta pagination (`page` dan `limit`).
- **Atomic Concurrency Stock Adjustment (Bonus):**
  - Endpoint `PATCH /items/:id/stock` dengan payload `delta`.
  - Dieksekusi melalui single-statement atomik SQL bersyarat (`WHERE stock + delta >= 0`). Menjamin stok tidak pernah bernilai negatif meskipun diakses oleh puluhan request simultan tanpa dependensi distributed lock/in-memory mutex.
- **Validasi Terpusat & Error Handling Standar:**
  - Pesan error validasi dalam Bahasa Indonesia dengan format JSON konsisten.
  - Pemetaan error domain terpusat (`400`, `401`, `404`, `409`, `422`, `500`).
- **DevOps & Dokumentasi:**
  - Siap dijalankan dengan satu perintah melalui Docker Compose.
  - Koleksi Postman lengkap (`docs/postman_collection.json`) dengan otomatisasi penyimpanan token.
  - Dokumentasi OpenAPI 3.0 (`docs/openapi.yaml`).

---

## 2. Tech Stack

- **Bahasa:** Go (Golang) 1.22+
- **HTTP Web Framework:** Gin (`github.com/gin-gonic/gin`)
- **ORM & Database:** GORM (`gorm.io/gorm`) & PostgreSQL 16 (`gorm.io/driver/postgres`)
- **Autentikasi & Kriptografi:** `golang-jwt/jwt/v5`, `golang.org/x/crypto/bcrypt`
- **Validasi:** `go-playground/validator/v10`
- **Migrasi Database:** `golang-migrate/migrate/v4` (embed SQL via Go `iofs`)
- **Testing:** `github.com/stretchr/testify`
- **Kontainerisasi:** Docker & Docker Compose

---

## 3. Struktur Folder & Arsitektur

Arsitektur yang diterapkan adalah **Layered Architecture** dengan dependensi satu arah dan pemisahan tanggung jawab (*Separation of Concerns*):

```
task-backend/
├── cmd/
│   └── api/
│       └── main.go                 # Entrypoint aplikasi, wiring DI manual, graceful shutdown
├── internal/
│   ├── config/
│   │   └── config.go               # Pembacaan dan validasi environment variables
│   ├── domain/                     # Entity inti dan sentinel error domain
│   │   ├── item.go
│   │   ├── user.go
│   │   └── errors.go
│   ├── dto/                        # Request dan response transfer objects + validation tags
│   │   ├── auth.go
│   │   ├── item.go
│   │   └── common.go
│   ├── handler/                    # HTTP presentation layer (binding, response, status code)
│   │   ├── auth_handler.go
│   │   ├── item_handler.go
│   │   └── health_handler.go
│   ├── service/                    # Business logic layer
│   │   ├── auth_service.go
│   │   └── item_service.go
│   ├── repository/                 # Data access layer (GORM queries & error mapping)
│   │   ├── user_repository.go
│   │   └── item_repository.go
│   ├── middleware/                 # RequestID, Logger, Recovery, CORS, Auth JWT
│   │   ├── auth.go
│   │   ├── cors.go
│   │   ├── logger.go
│   │   ├── recovery.go
│   │   └── request_id.go
│   ├── router/
│   │   └── router.go               # Konfigurasi engine Gin dan rute endpoint
│   ├── database/                   # Koneksi DB, connection pool, migrasi database
│   │   ├── database.go
│   │   └── migrate.go
│   └── pkg/                        # Utility internal reusable
│       ├── httpx/                  # Format standar response sukses & penanganan error
│       ├── password/               # Wrapper hashing bcrypt
│       └── token/                  # Wrapper penerbitan & verifikasi JWT
├── migrations/                     # File migrasi SQL murni (up dan down)
│   ├── 000001_create_users.up.sql
│   ├── 000001_create_users.down.sql
│   ├── 000002_create_items.up.sql
│   └── 000002_create_items.down.sql
├── docs/                           # Dokumentasi API
│   ├── postman_collection.json
│   └── openapi.yaml
├── test/
│   └── integration/                # Test konkurensi (tag integration)
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── .env.example
├── .golangci.yml
└── README.md
```

---

## 4. Cara Menjalankan

### Menggunakan Docker Compose (Direkomendasikan)

1. Salin berkas lingkungan:
   ```bash
   cp .env.example .env
   ```
2. Jalankan aplikasi dan database PostgreSQL:
   ```bash
   docker compose up --build
   ```
3. API akan aktif di `http://localhost:8080`.
   Migrasi database SQL akan otomatis berjalan saat kontainer API menyala.

### Menjalankan Lokal (Tanpa Docker untuk Aplikasi)

Pastikan PostgreSQL aktif di port 5432, lalu:
1. Konfigurasi file `.env` sesuai kredensial DB lokal.
2. Jalankan aplikasi:
   ```bash
   go run ./cmd/api
   ```

---

## 5. Menjalankan Pengujian (Testing)

- **Unit Test Service & Handler:**
  ```bash
  go test -v ./internal/...
  ```
- **Integration Test Konkurensi Stok:**
  ```bash
  go test -v -tags=integration ./test/integration/...
  ```

---

## 6. Contoh Pemanggilan API (cURL)

### 1. Registrasi Akun
```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Budi Santoso",
    "email": "budi@example.com",
    "password": "password123"
  }'
```

### 2. Login
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "budi@example.com",
    "password": "password123"
  }'
```
*Ambil token dari response `data.access_token`.*

### 3. Tambah Barang
```bash
curl -X POST http://localhost:8080/items \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <TOKEN_ANDA>" \
  -d '{
    "name": "Kabel USB-C 1m",
    "description": "Kabel data fast charging",
    "stock": 20,
    "price": 35000
  }'
```

### 4. Daftar Barang (Search & Pagination)
```bash
curl -X GET "http://localhost:8080/items?q=kabel&page=1&limit=10" \
  -H "Authorization: Bearer <TOKEN_ANDA>"
```

### 5. Pengurangan Stok Atomik (PATCH)
```bash
curl -X PATCH http://localhost:8080/items/<ITEM_ID>/stock \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <TOKEN_ANDA>" \
  -d '{
    "delta": -1
  }'
```

---

## 7. Jawaban Pertanyaan Pemahaman

### Pertanyaan 1: Jelaskan alur request dari masuk ke aplikasi hingga menghasilkan response, dan mengapa arsitektur aplikasi dipecah menjadi beberapa lapisan?

**Alur Request:**
1. **Router & Middleware:** Request HTTP masuk ke engine Gin. Middleware `RequestID` menyematkan ID pelacakan unik, `Logger` memulai timer durasi, `Recovery` memasang interceptor panic, `CORS` memvalidasi origin, dan `Auth` (pada rute `/items`) mengekstrak serta memverifikasi JWT Bearer token dari header `Authorization`. Jika token tidak valid, request langsung dihentikan dengan status `401 Unauthorized`.
2. **Handler Layer:** Handler menerima context Gin, melakukan deserialisasi (binding) payload JSON atau query parameters ke dalam DTO, memvalidasi tag aturan (`required`, `gte=0`, format email), serta mengekstrak identitas pengguna (`user_id`). Jika input tidak valid, handler mengembalikan error `422 Unprocessable Entity` atau `400 Bad Request`.
3. **Service Layer:** Menjalankan logika bisnis murni. Lapisan ini tidak mengetahui framework web (Gin). Contohnya: menormalkan email menjadi huruf kecil, melakukan hashing password dengan bcrypt, menyusun entity domain, dan menghitung pagination `total_pages`.
4. **Repository Layer:** Berinteraksi langsung dengan database melalui GORM. Mengeksekusi query SQL terparameterisasi (mencegah SQL Injection). Mengonversi error spesifik database (seperti duplikasi index unik PostgreSQL kode 23505 atau record not found) menjadi sentinel error domain (`ErrEmailTaken`, `ErrNotFound`).
5. **Response Translation:** Handler menerima entity atau error dari service, lalu memanggil fungsi pembantu di package `httpx` untuk menyusun amplop respons JSON standar (`data` / `error`) dan menentukan status HTTP yang tepat.

**Alasan Pemecahan Lapisan (Separation of Concerns):**
- **Single Responsibility Principle:** Setiap lapisan hanya memiliki satu alasan untuk berubah. Perubahan framework web (misalnya migrasi dari Gin ke Fiber atau standard library `net/http`) hanya menyentuh lapisan Handler tanpa mengganggu Service atau Repository.
- **Kemudahan Pengujian (Testability):** Logika bisnis di Service dapat diuji secara terisolasi menggunakan Mock Repository tanpa perlu menghidupkan server HTTP atau database nyata.
- **Maintainability & Skalabilitas:** Kode tersusun rapi dan mudah ditelusuri oleh anggota tim baru karena kontrak antarlapisan didefinisikan secara eksplisit menggunakan Go interface.

---

### Pertanyaan 2: Bagaimana pertimbangan keamanan dalam menyimpan token JWT di sisi client (frontend), antara HttpOnly Cookie vs LocalStorage?

1. **HttpOnly Cookie (dengan flag Secure dan SameSite=Strict/Lax):**
   - **Kelebihan:** Sangat tahan terhadap serangan **XSS (Cross-Site Scripting)** karena JavaScript di browser tidak memiliki akses baca/tulis ke cookie tersebut (`document.cookie` tidak bisa mengaksesnya).
   - **Kelemahan & Risiko:** Rentan terhadap serangan **CSRF (Cross-Site Request Forgery)** jika tidak dilindungi dengan benar. Diperlukan implementasi token Anti-CSRF (seperti pola double-submit cookie atau custom header) serta konfigurasi header `Access-Control-Allow-Credentials: true` pada CORS jika frontend dan backend berada di domain berbeda.
2. **LocalStorage:**
   - **Kelebihan:** Sangat mudah diimplementasikan pada Single Page Application (SPA). Frontend cukup menyematkan token secara manual ke header `Authorization: Bearer <token>`. Bebas dari risiko serangan CSRF murni karena browser tidak otomatis melampirkan isi LocalStorage pada request lintas domain.
   - **Kelemahan & Risiko:** Sangat rentan terhadap pencurian token jika terjadi celah **XSS**. Satu skrip jahat dari dependensi npm pihak ketiga (supply chain attack) atau injeksi konten pengguna yang tidak tersanitasi dapat mengeksekusi `localStorage.getItem('token')` dan mengirimkannya ke server penyerang.

**Kesimpulan Arsitektur:**
Untuk aplikasi dengan kebutuhan keamanan tinggi pada lingkungan produksi (perbankan/fintech), pendekatan **HttpOnly Secure SameSite Cookie** yang dilengkapi token Anti-CSRF adalah standar industri yang lebih aman. Pada API ini, token dikembalikan pada response body JSON sesuai instruksi soal untuk mempermudah integrasi klien API universal (Postman, mobile apps, SPA), dengan rekomendasi kepada pengembang frontend untuk meminimalisasi risiko celah XSS jika menyimpannya di browser.

---

### Pertanyaan 3: Bagaimana sistem mencegah terjadinya race condition dan stok negatif saat banyak request masuk bersamaan pada pengurangan stok barang?

Masalah umum pada pengurangan stok adalah kondisi *Read-Then-Write race condition*. Jika dua request membaca stok yang bernilai `1` pada saat bersamaan, keduanya akan mengevaluasi bahwa stok masih tersedia, lalu keduanya mengeksekusi pengurangan stok sebesar `1`, sehingga stok akhir menjadi `-1`.

Sistem ini menerapkan **solusi pertahanan berlapis** yang berfokus pada database atomicity:

1. **Single-Statement Atomic Conditional Update (Solusi Utama):**
   Pengurangan stok dieksekusi dalam satu statement SQL atomik dengan klausa kondisional langsung di database:
   ```sql
   UPDATE items
   SET stock = stock + $1, updated_at = now()
   WHERE id = $2 AND stock + $1 >= 0
   RETURNING id, name, description, stock, price, created_by, created_at, updated_at;
   ```
   - Database PostgreSQL menerapkan *row-level lock* eksklusif saat mengeksekusi operasi `UPDATE`.
   - Request pertama yang masuk akan mengunci baris barang, mengurangi stok (misalnya dari 1 menjadi 0), lalu melepaskan kunci.
   - Request kedua yang menunggu kunci kemudian mengevaluasi kondisi `WHERE stock + (-1) >= 0`. Karena stok saat itu sudah `0`, kondisi bernilai *false*, sehingga tidak ada baris yang terpengaruh (`RowsAffected == 0`).
   - Repository menangkap kondisi ini, memverifikasi keberadaan barang, lalu mengembalikan domain error `ErrInsufficientStock` yang dipetakan ke status HTTP `409 Conflict`.
2. **Database Integrity Constraint (Pertahanan Terakhir):**
   Pada tabel `items`, dibuat constraint eksplisit:
   ```sql
   CONSTRAINT chk_items_stock_non_negative CHECK (stock >= 0)
   ```
   Hal ini menjamin bahwa tidak ada transaksi apa pun yang dapat membuat nilai kolom `stock` berada di bawah nol, bahkan jika terdapat query manual dari luar aplikasi.
3. **Alternatif Pendekatan Lain:**
   - *Pessimistic Locking:* Menggunakan transaksi `BEGIN; SELECT * FROM items WHERE id = $1 FOR UPDATE; ... UPDATE ...; COMMIT;`. Efektif tetapi memerlukan transaksi terbuka lebih lama yang dapat mengurangi throughput database.
   - *Optimistic Locking:* Menambahkan kolom `version INTEGER`. Update dilakukan dengan mencocokkan nomor versi (`WHERE id = $1 AND version = $2`). Jika nomor versi bentrok, aplikasi melakukan retry.
4. **Pembuktian:**
   Diimplementasikan dalam test integrasi otomatis (`test/integration/concurrency_test.go`) di mana 50 goroutine dijalankan secara simultan untuk mengurangi stok barang yang awalnya bernilai 10. Hasil pengujian membuktikan tepat 10 goroutine berhasil, 40 goroutine menerima penolakan `409 Conflict`, dan stok akhir barang tepat bernilai 0.

