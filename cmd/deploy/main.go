// cmd/deploy/main.go — jalankan scraper secara lokal untuk testing
// Usage: go run ./cmd/deploy
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/yourusername/gcp-scraper-cron/internal/scraper"
)

func main() {
	ctx := context.Background()

	fmt.Println("🚀 Menjalankan scraper secara lokal...")

	results, err := scraper.Run(ctx)
	if err != nil {
		log.Fatalf("Scraper error: %v", err)
	}

	fmt.Printf("✅ Berhasil scrape %d item:\n\n", len(results))
	for _, item := range results {
		fmt.Printf("  [%s] %s: %s\n", item.Source, item.Title, item.Value)
	}

	// Set env var untuk test notifikasi Telegram
	// export TELEGRAM_TOKEN=xxx
	// export TELEGRAM_CHAT_ID=yyy
	if os.Getenv("TELEGRAM_TOKEN") != "" {
		fmt.Println("\n📨 Mengirim notifikasi Telegram...")
		// notifier.Send(ctx, results) — uncomment jika mau test
	}
}
