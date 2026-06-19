package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/yourusername/gcp-scraper-cron/internal/envloader"
	"github.com/yourusername/gcp-scraper-cron/internal/notifier"
	"github.com/yourusername/gcp-scraper-cron/internal/scraper"
	"github.com/yourusername/gcp-scraper-cron/internal/store"
)

func main() {
	envloader.Load(".env")

	ctx := context.Background()

	fmt.Println("Running scraper locally...")

	results, err := scraper.Run(ctx)
	if err != nil {
		log.Fatalf("scraper error: %v", err)
	}

	fmt.Printf("Scraped %d items:\n\n", len(results))
	for _, item := range results {
		fmt.Printf("  [%s] %s: %s\n", item.Source, item.Title, item.Value)
	}

	// Save to Firestore
	projectID := os.Getenv("GCP_PROJECT_ID")
	if projectID != "" {
		fmt.Println("\nSaving to Firestore...")
		db, err := store.New(ctx, projectID)
		if err != nil {
			log.Printf("firestore init error: %v", err)
		} else {
			defer db.Close()
			runID, err := db.SaveRun(ctx, results)
			if err != nil {
				log.Printf("firestore save error: %v", err)
			} else {
				fmt.Printf("Saved as run: %s\n", runID)
			}
		}
	} else {
		fmt.Println("\nSkipping Firestore (GCP_PROJECT_ID not set)")
	}

	// Send Telegram notification
	fmt.Println("\nSending Telegram notification...")
	if err := notifier.Send(ctx, results); err != nil {
		log.Fatalf("telegram error: %v", err)
	}
	fmt.Println("Done! Check your Telegram.")
}
