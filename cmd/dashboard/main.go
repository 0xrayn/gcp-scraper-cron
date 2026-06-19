// cmd/dashboard — a small read-only web UI showing scrape run history from Firestore.
// Deployed separately from the scraper itself, as a Cloud Run service.
//
// Routes:
//   GET /              list of recent runs
//   GET /runs/{run_id}  items from one specific run
package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/yourusername/gcp-scraper-cron/internal/envloader"
	"github.com/yourusername/gcp-scraper-cron/internal/store"
)

var (
	db         *store.Client
	listTmpl   = template.Must(template.New("list").Parse(listHTML))
	detailTmpl = template.Must(template.New("detail").Parse(detailHTML))
)

func main() {
	envloader.Load(".env")

	ctx := context.Background()

	projectID := os.Getenv("GCP_PROJECT_ID")
	if projectID == "" {
		log.Fatal("GCP_PROJECT_ID env var is required")
	}

	var err error
	db, err = store.New(ctx, projectID)
	if err != nil {
		log.Fatalf("firestore init failed: %v", err)
	}
	defer db.Close()

	http.HandleFunc("/", handleList)
	http.HandleFunc("/runs/", handleDetail)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Cloud Run sets PORT automatically; 8080 is the local fallback
	}

	log.Printf("dashboard listening on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func handleList(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	ctx := r.Context()
	runs, err := db.GetRecentRuns(ctx, 20)
	if err != nil {
		http.Error(w, "failed to load runs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	listTmpl.Execute(w, runs)
}

func handleDetail(w http.ResponseWriter, r *http.Request) {
	runID := strings.TrimPrefix(r.URL.Path, "/runs/")
	if runID == "" {
		http.NotFound(w, r)
		return
	}

	ctx := r.Context()
	items, err := db.GetRunItems(ctx, runID)
	if err != nil {
		http.Error(w, "failed to load items: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		RunID string
		Items any
	}{RunID: runID, Items: items}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	detailTmpl.Execute(w, data)
}
