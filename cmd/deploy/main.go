package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/yourusername/gcp-scraper-cron/internal/scraper"
)

func main() {
	// Load .env file kalau ada
	loadEnv(".env")

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
}

// loadEnv membaca file .env dan set ke os environment
func loadEnv(filename string) {
	f, err := os.Open(filename)
	if err != nil {
		return // .env tidak ada, skip
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// skip baris kosong dan komentar
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// hanya set jika belum ada di environment
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}
}
