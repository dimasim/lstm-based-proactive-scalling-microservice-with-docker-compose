# Testbed Skripsi — Arsitektur Microservices dengan Auto-Scaling Berbasis LSTM

Proyek ini merupakan testbed untuk penelitian skripsi yang mengkaji perbandingan beban komputasi antara dua alur autentikasi berbasis JWT: **proses login (CPU-heavy)** pada Auth Service (NestJS) vs **verifikasi token (stateless/ringan)** pada Quiz Service (Go). Data telemetri yang dikumpulkan digunakan sebagai input multivariate untuk model prediksi auto-scaling berbasis LSTM.

---

## Arsitektur Sistem

```
                        ┌─────────────┐
                        │   Locust    │  Load Generator
                        └──────┬──────┘
                               │ HTTP
                        ┌──────▼──────┐
                        │   HAProxy   │  :80  (reverse proxy & load balancer)
                        │             │  :8404 (metrics & stats)
                        └──────┬──────┘
               ┌───────────────┴───────────────┐
               │ /auth/*                        │ /quiz
      ┌────────▼────────┐             ┌─────────▼────────┐
      │  Auth Service   │             │  Quiz Service    │
      │  (NestJS :3000) │             │  (Go :8080)      │
      └────────┬────────┘             └─────────┬────────┘
               │                               │
               └──────────────┬────────────────┘
                        ┌──────▼──────┐
                        │ PostgreSQL  │  auth_db | quiz_db
                        └─────────────┘

Monitoring Stack:
  cAdvisor :8088 → Prometheus :9090 → Grafana :3000
  HAProxy  :8404 → Prometheus :9090
```

---

## Layanan

| Service | Image / Stack | Port | Deskripsi |
|---|---|---|---|
| `auth-service` | NestJS + bcrypt | 3000 | Login, password hashing, JWT signing |
| `quiz-service` | Go (net/http) | 8080 | Verifikasi JWT, logging quiz attempt |
| `haproxy` | HAProxy 2.8 | 80, 8404 | Reverse proxy, routing, metrics exporter |
| `postgres` | PostgreSQL 15 | 5432 | `auth_db` (users) & `quiz_db` (quiz_attempts) |
| `cadvisor` | cAdvisor v0.47 | 8088 | Container resource metrics (CPU, RAM) |
| `prometheus` | Prometheus 2.45 | 9090 | Scrape & store telemetry (retensi 3 hari) |
| `grafana` | Grafana 10.0 | 3000 | Dashboard visualisasi real-time |
| `locust` | Locust | 8089 | Load testing berbasis dataset CSV |

---

## Struktur Direktori

```
skripsi/
├── auth-service/              # NestJS Auth Service
│   ├── src/
│   │   ├── main.ts
│   │   ├── app.module.ts
│   │   ├── app.controller.ts  # POST /auth/login
│   │   └── app.service.ts     # DB, bcrypt, JWT signing
│   ├── Dockerfile
│   └── package.json
│
├── quiz-service/              # Go Quiz Service
│   ├── main.go                # Entry point & wiring
│   ├── config/config.go       # Env vars (DATABASE_URL, PORT, JWTSecret)
│   ├── database/database.go   # Koneksi DB + migrasi DDL
│   ├── middleware/jwt.go      # JWT Bearer validation
│   ├── handler/quiz.go        # POST /quiz, GET /healthz
│   ├── model/model.go         # Struct Claims
│   └── go.mod
│
├── haproxy/haproxy.cfg        # Routing rules & stats endpoint
├── prometheus/prometheus.yml  # Scrape config (cadvisor, haproxy)
├── grafana/
│   ├── dashboards/            # Dashboard JSON
│   └── provisioning/          # Auto-provisioning datasource & dashboard
├── postgres/
│   └── init-multiple-databases.sh  # Init auth_db & quiz_db
├── locust/locustfile.py       # Skenario load test (login + quiz)
├── docker-compose.yml
└── README.md
```

---

## Instalasi di Server Linux (Fresh Install)

Bagian ini untuk server Linux (Ubuntu/Debian) yang belum terinstall apapun.

### 1. Update sistem

```bash
sudo apt update && sudo apt upgrade -y
```

### 2. Install Docker Engine

```bash
# Install dependensi
sudo apt install -y ca-certificates curl gnupg lsb-release

# Tambah GPG key & repository Docker
sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
sudo chmod a+r /etc/apt/keyrings/docker.gpg

echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] \
  https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

# Install Docker
sudo apt update
sudo apt install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
```

Verifikasi:

```bash
docker --version        # Docker version 24.x.x
docker compose version  # Docker Compose version v2.x.x
```

### 3. Jalankan Docker tanpa sudo (opsional tapi direkomendasikan)

```bash
sudo usermod -aG docker $USER
newgrp docker
```

### 4. Install Git (untuk clone repo)

```bash
sudo apt install -y git
```

### 5. Clone repositori

```bash
git clone <url-repo> skripsi
cd skripsi
```

> Tidak perlu install Node.js, Go, atau runtime lain — semuanya sudah dikemas di dalam Docker image masing-masing service.

---

## Prasyarat

- [Docker](https://docs.docker.com/get-docker/) >= 24
- [Docker Compose](https://docs.docker.com/compose/) >= 2.20
- File dataset CSV (lihat bagian [Dataset](#dataset))

---

## Menjalankan Proyek

### 1. Siapkan file dataset

Pastikan file berikut ada di root direktori sebelum menjalankan compose:

```
skripsi/
├── ISTN101_unique_users.csv     # Kolom: Student_ID
├── ISTN101_quiz_attempts.csv    # Kolom: Student_ID
└── list_login_user_diawal.csv   # Kolom: Student_ID
```

### 2. Build & jalankan semua layanan

```bash
docker compose up --build -d
```

### 3. Cek status layanan

```bash
docker compose ps
```

Semua layanan harus berstatus `running`. Tunggu ±10–15 detik untuk PostgreSQL selesai inisialisasi sebelum service lain siap menerima request.

### 4. Akses UI

| UI | URL | Kredensial |
|---|---|---|
| Grafana | http://localhost:3000 | `admin` / `admin` |
| Prometheus | http://localhost:9090 | — |
| HAProxy Stats | http://localhost:8404/stats | — |
| Locust | http://localhost:8089 | — |
| cAdvisor | http://localhost:8088 | — |

### 5. Jalankan load test

Buka http://localhost:8089, set jumlah user dan spawn rate, lalu klik **Start Swarming**.

Atau via CLI:

```bash
docker compose run --rm locust \
  -f /mnt/locust/locustfile.py \
  --host http://haproxy \
  --users 50 \
  --spawn-rate 5 \
  --run-time 5m \
  --headless
```

### 6. Hentikan semua layanan

```bash
docker compose down
```

Untuk menghapus data PostgreSQL juga:

```bash
docker compose down -v
```

---

## Alur Request

### Kasus 1 — Login (CPU-heavy di NestJS)

```
Locust → HAProxy /auth/login → auth-service
           POST { student_id, password }
           ↓
           DB query → bcrypt.compare() → jwtService.sign()
           ↓
           { access_token: "eyJ..." }
```

Proses ini memicu CPU usage tinggi karena bcrypt hashing (cost factor 10).

### Kasus 2 — Quiz (stateless di Go)

```
Locust → HAProxy /quiz → quiz-service
           GET, Header: Authorization: Bearer <token>
           ↓
           jwt.ParseWithClaims() [in-memory]
           ↓
           INSERT INTO quiz_attempts
           ↓
           { status: "success", student_id: "..." }
```

Verifikasi JWT hanya operasi HMAC di memori, sehingga CPU usage Go tetap rendah meskipun RPS tinggi.

---

## Telemetri & Metrik

Prometheus meng-scrape metrik setiap **5 detik** dari dua sumber:

| Source | Target | Metrik Utama |
|---|---|---|
| cAdvisor | `cadvisor:8080` | CPU usage, RAM (working set) per container |
| HAProxy | `haproxy:8404` | RPS per backend (`auth_back`, `quiz_back`) |

### 8 Fitur Telemetri (Multivariate Input Model)

| # | Nama | Sumber |
|---|---|---|
| 1 | `go_cpu_usage` | cAdvisor — container `quiz-service` |
| 2 | `go_ram_usage` | cAdvisor — container `quiz-service` |
| 3 | `go_rps` | HAProxy — backend `quiz_back` |
| 4 | `go_current_container` | Docker API via cAdvisor |
| 5 | `nestjs_cpu_usage` | cAdvisor — container `auth-service` |
| 6 | `nestjs_ram_usage` | cAdvisor — container `auth-service` |
| 7 | `nestjs_rps` | HAProxy — backend `auth_back` |
| 8 | `nestjs_current_container` | Docker API via cAdvisor |

---

## Validasi Sistem Berhasil

Sistem dianggap siap mengumpulkan data jika:

1. Semua target di Prometheus (`http://localhost:9090/targets`) berstatus **UP**.
2. Dashboard Grafana menampilkan data real-time yang berfluktuasi sesuai beban Locust.
3. **Panel Perbandingan Kriptografi** — RPS login NestJS naik → CPU NestJS naik tajam.
4. **Panel Efisiensi Verifikasi** — RPS quiz Go tinggi → CPU NestJS tetap flat.

---

## Environment Variables

### auth-service

| Variabel | Default | Keterangan |
|---|---|---|
| `DATABASE_URL` | `postgres://postgres:postgres@postgres:5432/auth_db?sslmode=disable` | Koneksi PostgreSQL |
| `PORT` | `3000` | Port HTTP server |

### quiz-service

| Variabel | Default | Keterangan |
|---|---|---|
| `DATABASE_URL` | `postgres://postgres:postgres@postgres:5432/quiz_db?sslmode=disable` | Koneksi PostgreSQL |
| `PORT` | `8080` | Port HTTP server |

---

## API Endpoints

### Auth Service (`/auth`)

```
POST /auth/login
Content-Type: application/json

{ "student_id": "STUD0014859", "password": "password123" }

→ 200 { "access_token": "eyJ..." }
→ 401 { "message": "Invalid credentials" }
```

### Quiz Service (`/quiz`)

```
POST /quiz
Authorization: Bearer <token>
Content-Type: application/json

{
  "question_id": 1,
  "selected_option": "Central Processing Unit"
}

→ 200 { 
        "status": "success", 
        "correct": true, 
        "next_question": { 
          "id": 2, 
          "text": "...", 
          "options": [...] 
        } 
      }
→ 400 { "status": "error", "message": "Invalid JSON payload" }
→ 401 { "error": "Invalid token" }

GET /healthz
→ 200 OK
```

---

## Dataset

File CSV yang digunakan berasal dari dataset aktivitas mahasiswa kursus **ISTN101**:

- `ISTN101_unique_users.csv` — daftar Student ID unik untuk seeding database dan login awal.
- `ISTN101_quiz_attempts.csv` — log percobaan kuis sebagai pola beban kerja sekuensial.
- `list_login_user_diawal.csv` — subset user yang langsung login saat Locust start.

---

## Referensi

- [HAProxy Prometheus Exporter](https://www.haproxy.com/blog/haproxy-exposes-a-prometheus-metrics-endpoint)
- [cAdvisor](https://github.com/google/cadvisor)
- [golang-jwt/jwt](https://github.com/golang-jwt/jwt)
- [Locust Documentation](https://docs.locust.io)
