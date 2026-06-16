package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/yourusername/gcp-scraper-cron/internal/notifier"
	"github.com/yourusername/gcp-scraper-cron/internal/scraper"
)

func init() {
	// Register HTTP function — Cloud Scheduler akan hit endpoint ini
	functions.HTTP("RunScraper", RunScraper)
}

// RunScraper adalah entry point Cloud Function.
// Dipanggil otomatis oleh Cloud Scheduler sesuai jadwal cron.
func RunScraper(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Validasi request dari Cloud Scheduler (opsional tapi direkomendasikan)
	if err := validateSchedulerRequest(r); err != nil {
		log.Printf("unauthorized request: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	log.Println("Scraper dimulai:", time.Now().Format(time.RFC3339))

	// Jalankan scraper
	results, err := scraper.Run(ctx)
	if err != nil {
		log.Printf("Scraper error: %v", err)
		http.Error(w, fmt.Sprintf("Scraper gagal: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Berhasil scrape %d item", len(results))

	// Kirim notifikasi (Telegram / Email)
	if err := notifier.Send(ctx, results); err != nil {
		// Notif gagal tidak menghentikan proses — hanya log
		log.Printf("Notifikasi gagal: %v", err)
	}

	// Response sukses
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status":  "ok",
		"scraped": len(results),
		"time":    time.Now().Format(time.RFC3339),
	})
}

// validateSchedulerRequest memastikan request berasal dari Cloud Scheduler
// dengan memeriksa header khusus yang kita set di konfigurasi scheduler.
func validateSchedulerRequest(r *http.Request) error {
	secret := os.Getenv("SCHEDULER_SECRET")
	if secret == "" {
		// Jika env tidak di-set, skip validasi (untuk development)
		return nil
	}

	authHeader := r.Header.Get("X-Scheduler-Secret")
	if authHeader != secret {
		return fmt.Errorf("invalid scheduler secret")
	}
	return nil
}

// HealthCheck endpoint untuk monitoring — bisa ditambahkan sebagai function terpisah
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy"}`))
}
