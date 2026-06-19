package scraper

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// scrapeHTMLTarget fetches an HTML page and extracts repeating items using
// CSS selectors defined in the target config. Useful for product listings,
// news headlines, job boards — any page with a repeating card/row structure.
func scrapeHTMLTarget(ctx context.Context, t Target) ([]Item, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, t.URL, nil)
	if err != nil {
		return nil, err
	}
	// Many sites block requests without a normal browser User-Agent
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; ScraperBot/1.0)")

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parse HTML: %w", err)
	}

	if t.ItemSelector == "" {
		return nil, fmt.Errorf("target %q: item_selector is required for html type", t.Name)
	}

	now := time.Now()
	var items []Item

	doc.Find(t.ItemSelector).EachWithBreak(func(i int, s *goquery.Selection) bool {
		if t.MaxItems > 0 && i >= t.MaxItems {
			return false // stop iterating
		}

		title := strings.TrimSpace(s.Find(t.TitleSelector).First().Text())

		var value string
		valueEl := s.Find(t.ValueSelector).First()
		if t.ValueAttr != "" {
			// Extract an attribute (e.g. href) instead of text content
			attrVal, exists := valueEl.Attr(t.ValueAttr)
			if exists {
				value = strings.TrimSpace(attrVal)
			}
		} else {
			value = strings.TrimSpace(valueEl.Text())
		}

		// Skip items where both selectors came up empty — usually means
		// the selector doesn't match anything on the page (site changed structure)
		if title == "" && value == "" {
			return true
		}

		items = append(items, Item{
			Title:     title,
			Value:     value,
			Source:    t.Source,
			ScrapedAt: now,
		})
		return true
	})

	if len(items) == 0 {
		return nil, fmt.Errorf("no items matched selector %q on target %q — page structure may have changed", t.ItemSelector, t.Name)
	}

	return items, nil
}
