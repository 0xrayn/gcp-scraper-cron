package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// scrapeJSONTarget fetches a JSON API and extracts fields defined in the target config.
// It supports simple dot-paths like "rates.IDR" to reach nested values.
func scrapeJSONTarget(ctx context.Context, t Target) ([]Item, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, t.URL, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var raw map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode JSON: %w", err)
	}

	format := t.JSONValueFormat
	if format == "" {
		format = "%v"
	}

	now := time.Now()
	var items []Item

	for title, path := range t.JSONFields {
		val, ok := lookupPath(raw, path)
		if !ok {
			continue // field missing in response — skip rather than fail the whole target
		}
		items = append(items, Item{
			Title:     title,
			Value:     fmt.Sprintf(format, val),
			Source:    t.Source,
			ScrapedAt: now,
		})
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("no fields matched in response for target %q", t.Name)
	}

	return items, nil
}

// lookupPath walks a dot-separated path ("rates.IDR") through nested maps.
func lookupPath(data map[string]any, path string) (any, bool) {
	parts := strings.Split(path, ".")
	var current any = data

	for _, part := range parts {
		m, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		val, exists := m[part]
		if !exists {
			return nil, false
		}
		current = val
	}

	return current, true
}
