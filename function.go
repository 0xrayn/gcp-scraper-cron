package scraper

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/yourusername/gcp-scraper-cron/internal/notifier"
	"github.com/yourusername/gcp-scraper-cron/internal/scraper"
	"github.com/yourusername/gcp-scraper-cron/internal/store"
)

func init() {
	functions.HTTP("RunScraper", RunScraper)
}

func RunScraper(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := validateSchedulerRequest(r); err != nil {
		log.Printf("unauthorized request: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	log.Println("Scraper started:", time.Now().Format(time.RFC3339))

	// Init Firestore
	projectID := os.Getenv("GCP_PROJECT_ID")
	db, err := store.New(ctx, projectID)
	if err != nil {
		log.Printf("firestore init error: %v", err)
		http.Error(w, "storage unavailable", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Run scraper
	results, err := scraper.Run(ctx)
	if err != nil {
		log.Printf("scraper error: %v", err)
		db.SaveFailedRun(ctx, err)
		http.Error(w, fmt.Sprintf("scraper failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Save to Firestore
	runID, err := db.SaveRun(ctx, results)
	if err != nil {
		// Log but don't stop — still send notification
		log.Printf("firestore save error: %v", err)
	} else {
		log.Printf("saved run %s with %d items", runID, len(results))
	}

	// Send Telegram notification
	if err := notifier.Send(ctx, results); err != nil {
		log.Printf("notification failed: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status":  "ok",
		"run_id":  runID,
		"scraped": len(results),
		"time":    time.Now().Format(time.RFC3339),
	})
}

func validateSchedulerRequest(r *http.Request) error {
	secret := os.Getenv("SCHEDULER_SECRET")
	if secret == "" {
		return nil
	}
	if r.Header.Get("X-Scheduler-Secret") != secret {
		return fmt.Errorf("invalid scheduler secret")
	}
	return nil
}
