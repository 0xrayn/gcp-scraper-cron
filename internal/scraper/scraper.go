// Package scraper runs a set of scrape targets defined in a config file.
// Each target is either a JSON API call or an HTML page scrape — see
// config.go for the target schema and configs/targets.json for examples.
package scraper

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
)

// Item is one scraped result, regardless of which target produced it.
type Item struct {
	Title     string    `json:"title"`
	Value     string    `json:"value"`
	Source    string    `json:"source"`
	ScrapedAt time.Time `json:"scraped_at"`
}

// defaultConfigPath is where Run looks for the target config by default.
// Overridable via the SCRAPER_CONFIG_PATH env var (useful for tests or
// alternate deployments).
const defaultConfigPath = "configs/targets.json"

// Run loads the target config and scrapes every enabled target.
// A failure in one target is logged and skipped — it does not stop
// the other targets from running.
func Run(ctx context.Context) ([]Item, error) {
	configPath := os.Getenv("SCRAPER_CONFIG_PATH")
	if configPath == "" {
		configPath = defaultConfigPath
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	targets := cfg.EnabledTargets()
	if len(targets) == 0 {
		return nil, fmt.Errorf("no enabled targets in %s", configPath)
	}

	var allItems []Item
	var failedTargets []string

	for _, target := range targets {
		items, err := runTarget(ctx, target)
		if err != nil {
			log.Printf("target %q failed: %v", target.Name, err)
			failedTargets = append(failedTargets, target.Name)
			continue // keep going — one bad target shouldn't kill the whole run
		}
		allItems = append(allItems, items...)
	}

	if len(allItems) == 0 {
		return nil, fmt.Errorf("all %d targets failed: %v", len(targets), failedTargets)
	}

	if len(failedTargets) > 0 {
		log.Printf("completed with %d/%d targets failing: %v", len(failedTargets), len(targets), failedTargets)
	}

	return allItems, nil
}

// runTarget dispatches to the right scraping strategy based on target type.
func runTarget(ctx context.Context, t Target) ([]Item, error) {
	switch t.Type {
	case TypeJSON:
		return scrapeJSONTarget(ctx, t)
	case TypeHTML:
		return scrapeHTMLTarget(ctx, t)
	default:
		return nil, fmt.Errorf("unknown target type %q", t.Type)
	}
}
