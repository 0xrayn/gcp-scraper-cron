# 🕷️ GCP Scraper + Cron Job dengan Go

> Aplikasi scraper data otomatis yang berjalan terjadwal setiap hari di Google Cloud Platform, tanpa perlu server sendiri, tanpa perlu buka laptop — dan **gratis**.

---

## 📌 Apa Itu Aplikasi Ini?

Aplikasi ini adalah **scraper otomatis berbasis cloud** yang ditulis dengan bahasa Go. Program ini akan berjalan sendiri setiap hari sesuai jadwal yang kamu tentukan, mengambil data dari internet, lalu mengirimkan hasilnya langsung ke **Telegram** kamu.

Tidak perlu VPS yang nyala 24 jam. Tidak perlu buka laptop. Semua berjalan otomatis di Google Cloud.

---

## 🎯 Tujuan & Kegunaan

Aplikasi ini cocok digunakan untuk:

| Use Case | Contoh |
|---|---|
| 📈 Pantau harga produk | Cek harga laptop di Tokopedia/Shopee tiap pagi, notif kalau turun |
| 💱 Monitor kurs mata uang | Rekap USD/IDR/EUR/SGD setiap hari jam 07:00 WIB |
| 📰 Rangkuman berita | Ambil headline berita terbaru dari portal tertentu |
| 🛒 Cek stok barang | Alert otomatis kalau barang yang habis sudah kembali tersedia |
| 💼 Monitor lowongan kerja | Scrape job portal, notif kalau ada posisi baru yang sesuai |
| 📊 Laporan data berkala | Rekap data apapun dari web dan kirim ke tim via Telegram |

Selain itu, project ini juga bagus sebagai **portofolio GitHub** karena menunjukkan kemampuan:
- Menulis aplikasi Go yang bersih dan terstruktur
- Deploy ke Google Cloud Platform (Cloud Functions + Cloud Scheduler)
- Implementasi CI/CD otomatis dengan GitHub Actions
- Arsitektur serverless yang efisien dan hemat biaya

---

## 🏗️ Arsitektur Aplikasi

```
┌─────────────────────────────────────────────────────────┐
│                     GITHUB                              │
│  Repository ──(push ke main)──► GitHub Actions          │
│                                      │                  │
│                                      │ auto deploy      │
└──────────────────────────────────────┼──────────────────┘
                                       │
                                       ▼
┌─────────────────────────────────────────────────────────┐
│                  GOOGLE CLOUD PLATFORM                  │
│                                                         │
│  Cloud Scheduler                                        │
│  (cron: tiap hari 07:00 WIB)                           │
│        │                                                │
│        │ HTTP GET + Secret Header                       │
│        ▼                                                │
│  Cloud Functions Gen2 (Go 1.21)                        │
│        │                                                │
│        ├── internal/scraper ──► Ambil data dari web     │
│        │                                                │
│        └── internal/notifier ──► Kirim ke Telegram      │
└─────────────────────────────────────────────────────────┘
                                       │
                                       ▼
                              📱 Telegram Bot
                         (pesan masuk ke HP kamu)
```

**Alur kerja singkat:**
1. Cloud Scheduler "membangunkan" Cloud Function sesuai jadwal cron
2. Cloud Function menjalankan kode Go yang scrape data dari internet
3. Hasil scrape diformat dan dikirim ke Telegram bot kamu
4. Selesai — Cloud Function mati lagi (tidak makan biaya saat idle)

---

## 💰 Estimasi Biaya GCP (per bulan)

| Service | Penggunaan | Free Tier | Biaya |
|---|---|---|---|
| Cloud Functions Gen2 | 30 invocations/bulan | 2 juta/bulan | **GRATIS** |
| Cloud Scheduler | 1 job | 3 job gratis | **GRATIS** |
| Egress network | < 100 KB/hari | 1 GB/bulan | **GRATIS** |
| Cloud Build (deploy) | ~5 menit/deploy | 120 menit/hari | **GRATIS** |
| **Total** | | | **~$0/bulan** |

> ⚠️ Tetap aktifkan **Budget Alerts** di GCP Console (set limit $5) agar tidak kaget jika ada penggunaan tidak terduga.

---

## 📁 Struktur Folder

```
gcp-scraper-cron/
│
├── function.go                  # Entry point Cloud Function
│                                # Menerima HTTP request dari Cloud Scheduler
│
├── go.mod                       # Dependency Go
│
├── internal/
│   ├── scraper/
│   │   └── scraper.go           # Logika utama scraping data
│   │                            # Edit file ini untuk ganti/tambah sumber data
│   │
│   └── notifier/
│       └── telegram.go          # Kirim hasil ke Telegram Bot
│
├── cmd/
│   └── deploy/
│       └── main.go              # Runner lokal — untuk test tanpa deploy ke GCP
│
└── .github/
    └── workflows/
        └── deploy.yml           # CI/CD: auto deploy ke GCP setiap push ke main
```

---

## 🚀 Cara Setup & Deploy (Pertama Kali)

### Prasyarat

- Akun Google Cloud Platform (gratis daftar, ada $300 free credit)
- Akun GitHub
- [gcloud CLI](https://cloud.google.com/sdk/docs/install) terinstall di laptop
- Go 1.21+ terinstall
- Akun Telegram

---

### Langkah 1 — Buat Project GCP & Aktifkan API

```bash
# Login ke GCP
gcloud auth login

# Buat project baru (ganti nama sesuai keinginan)
gcloud projects create scraper-project-xxx
gcloud config set project scraper-project-xxx

# Aktifkan API yang dibutuhkan
gcloud services enable cloudfunctions.googleapis.com
gcloud services enable cloudscheduler.googleapis.com
gcloud services enable cloudbuild.googleapis.com
gcloud services enable run.googleapis.com
```

---

### Langkah 2 — Buat Service Account untuk GitHub Actions

Service Account ini dipakai oleh GitHub Actions agar bisa deploy ke GCP secara otomatis.

```bash
# Buat service account
gcloud iam service-accounts create github-deployer \
  --display-name="GitHub Actions Deployer"

# Berikan izin deploy Cloud Functions
gcloud projects add-iam-policy-binding scraper-project-xxx \
  --member="serviceAccount:github-deployer@scraper-project-xxx.iam.gserviceaccount.com" \
  --role="roles/cloudfunctions.developer"

# Berikan izin kelola Cloud Scheduler
gcloud projects add-iam-policy-binding scraper-project-xxx \
  --member="serviceAccount:github-deployer@scraper-project-xxx.iam.gserviceaccount.com" \
  --role="roles/cloudscheduler.admin"

# Berikan izin Cloud Run (dibutuhkan Cloud Functions Gen2)
gcloud projects add-iam-policy-binding scraper-project-xxx \
  --member="serviceAccount:github-deployer@scraper-project-xxx.iam.gserviceaccount.com" \
  --role="roles/run.admin"

# Berikan izin impersonate service account
gcloud projects add-iam-policy-binding scraper-project-xxx \
  --member="serviceAccount:github-deployer@scraper-project-xxx.iam.gserviceaccount.com" \
  --role="roles/iam.serviceAccountUser"

# Export key JSON — OUTPUT INI YANG DISIMPAN KE GITHUB SECRETS
gcloud iam service-accounts keys create key.json \
  --iam-account=github-deployer@scraper-project-xxx.iam.gserviceaccount.com

cat key.json  # copy seluruh isi file ini
```

---

### Langkah 3 — Buat Telegram Bot

1. Buka Telegram, cari **@BotFather**
2. Kirim perintah `/newbot`
3. Ikuti instruksi → masukkan nama bot → dapat **TOKEN** (simpan!)
4. Start bot kamu: cari nama bot di Telegram, klik **Start**
5. Untuk dapat **CHAT_ID**, buka browser:
   ```
   https://api.telegram.org/bot<TOKEN_KAMU>/getUpdates
   ```
6. Kirim pesan apapun ke bot kamu dulu, lalu refresh URL di atas
7. Cari nilai `"chat":{"id": 123456789}` — angka itu adalah **CHAT_ID** kamu

---

### Langkah 4 — Set GitHub Secrets

Buka repo GitHub → **Settings** → **Secrets and variables** → **Actions** → **New repository secret**

Tambahkan 5 secrets berikut:

| Secret Name | Nilai | Keterangan |
|---|---|---|
| `GCP_SA_KEY` | Isi seluruh file `key.json` | Kredensial Service Account |
| `GCP_PROJECT_ID` | `scraper-project-xxx` | ID project GCP kamu |
| `TELEGRAM_TOKEN` | Token dari BotFather | Untuk kirim pesan |
| `TELEGRAM_CHAT_ID` | Angka ID dari getUpdates | Target penerima pesan |
| `SCHEDULER_SECRET` | Password bebas, misal `rahasia123` | Keamanan endpoint Function |

---

### Langkah 5 — Push ke GitHub → Deploy Otomatis!

```bash
git init
git add .
git commit -m "feat: initial scraper setup"
git branch -M main
git remote add origin https://github.com/username/gcp-scraper-cron.git
git push -u origin main
```

GitHub Actions akan otomatis:
1. Menjalankan tests
2. Deploy Cloud Function ke GCP
3. Membuat/update Cloud Scheduler job

Cek progress deploy di tab **Actions** di repo GitHub kamu.

---

## 🖥️ Cara Menjalankan Secara Lokal (Testing)

### Test scraper langsung

```bash
# Clone repo
git clone https://github.com/username/gcp-scraper-cron.git
cd gcp-scraper-cron

# Install dependencies
go mod tidy

# Jalankan scraper — hasil tampil di terminal
go run ./cmd/deploy
```

### Simulasi Cloud Function di lokal

```bash
# Jalankan sebagai HTTP server lokal
go run github.com/GoogleCloudPlatform/functions-framework-go/funcframework \
  --target=RunScraper --port=8080

# Di terminal lain, panggil seperti Cloud Scheduler memanggil
curl http://localhost:8080
```

### Test dengan Telegram (lokal)

```bash
# Set env variable dulu
export TELEGRAM_TOKEN=token_dari_botfather
export TELEGRAM_CHAT_ID=chat_id_kamu

# Jalankan
go run ./cmd/deploy
```

---

## ✏️ Cara Menambahkan Scraper Baru

Edit file `internal/scraper/scraper.go`:

**1. Tambahkan fungsi scraper baru:**

```go
// Contoh: scrape harga BBM dari website SPBU
func scrapeHargaBBM(ctx context.Context) ([]Item, error) {
    resp, err := http.Get("https://mypertamina.id/harga-bbm")
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    // parsing HTML dengan goquery atau string manipulation
    // ...

    return []Item{
        {
            Title:     "Pertalite",
            Value:     "Rp 10.000/liter",
            Source:    "pertamina.com",
            ScrapedAt: time.Now(),
        },
    }, nil
}
```

**2. Daftarkan di fungsi `Run()`:**

```go
func Run(ctx context.Context) ([]Item, error) {
    var results []Item

    // Scraper yang sudah ada
    currencyItems, err := scrapeCurrency(ctx)
    results = append(results, currencyItems...)

    // Tambahkan scraper baru di sini
    bbmItems, err := scrapeHargaBBM(ctx)
    if err != nil {
        log.Printf("scraper BBM gagal: %v", err) // log saja, jangan stop semua
    } else {
        results = append(results, bbmItems...)
    }

    return results, nil
}
```

---

## ⏰ Mengubah Jadwal Cron

Edit bagian `--schedule` di `.github/workflows/deploy.yml`:

```yaml
--schedule="0 0 * * *"
```

| Format Cron | Artinya |
|---|---|
| `0 0 * * *` | Setiap hari jam 00:00 UTC (07:00 WIB) |
| `0 22 * * *` | Setiap hari jam 22:00 UTC (05:00 WIB pagi) |
| `0 */6 * * *` | Setiap 6 jam sekali |
| `0 0 * * 1` | Setiap Senin jam 07:00 WIB |
| `*/30 * * * *` | Setiap 30 menit |
| `0 0 1 * *` | Setiap tanggal 1 (awal bulan) |

> 💡 GCP Cloud Scheduler menggunakan **UTC**. Indonesia WIB = UTC+7, jadi kurangi 7 jam dari waktu yang kamu inginkan.

---

## 🔧 Troubleshooting

**Deploy gagal di GitHub Actions:**
- Pastikan semua 5 GitHub Secrets sudah diisi dengan benar
- Cek tab Actions → klik workflow yang gagal → baca error message-nya
- Pastikan API GCP sudah diaktifkan di Langkah 1

**Telegram tidak dapat pesan:**
- Pastikan kamu sudah pernah kirim pesan ke bot (bot tidak bisa mulai duluan)
- Cek TELEGRAM_TOKEN dan TELEGRAM_CHAT_ID sudah benar
- Test dengan curl:
  ```bash
  curl "https://api.telegram.org/bot<TOKEN>/getMe"
  ```

**Cloud Function timeout:**
- Default timeout 300 detik. Jika scraping banyak halaman, pertimbangkan pecah jadi beberapa function
- Tambahkan `--timeout=540s` di deploy command (max 9 menit untuk Gen2)

**Melihat logs Cloud Function:**
```bash
gcloud functions logs read scraper-harian \
  --region=asia-southeast2 \
  --limit=50
```

---

## 📚 Teknologi yang Digunakan

| Teknologi | Versi | Kegunaan |
|---|---|---|
| Go | 1.21 | Bahasa pemrograman utama |
| GCP Cloud Functions Gen2 | - | Serverless compute (menjalankan kode) |
| GCP Cloud Scheduler | - | Cron job (menjadwalkan eksekusi) |
| GitHub Actions | - | CI/CD (auto deploy) |
| Telegram Bot API | - | Notifikasi hasil scraping |
| functions-framework-go | v1.8 | Framework Cloud Functions untuk Go |

---

## 📄 Lisensi

MIT License — bebas digunakan dan dimodifikasi.
