// Package scraper — config.go defines the structure for scrape targets
// and loads them from a JSON config file. This makes the scraper
// "config-driven": adding a new target means editing configs/targets.json,
// not writing new Go code (unless the target needs custom logic).
package scraper

import (
	"encoding/json"
	"fmt"
	"os"
)

// TargetType decides which scraping strategy to use.
type TargetType string

const (
	TypeJSON TargetType = "json" // hit a JSON API, extract fields by path
	TypeHTML TargetType = "html" // fetch HTML, extract via CSS selectors
)

// Target describes one thing to scrape — one row in the config file.
type Target struct {
	// Name is a human label, e.g. "Kurs USD-IDR" or "Harga Laptop Tokopedia"
	Name string `json:"name"`

	// Enabled lets you turn a target on/off without deleting it from config
	Enabled bool `json:"enabled"`

	// Type selects the scraping strategy: "json" or "html"
	Type TargetType `json:"type"`

	// URL is the endpoint to fetch
	URL string `json:"url"`

	// Source is a short label shown in notifications, e.g. "tokopedia.com"
	Source string `json:"source"`

	// --- Fields used when Type == "json" ---
	// JSONFields maps an output title to a dot-path inside the JSON response.
	// Example: {"USD to IDR": "rates.IDR"} reads response["rates"]["IDR"].
	JSONFields map[string]string `json:"json_fields,omitempty"`

	// JSONValueFormat is a Printf-style format applied to the extracted value.
	// Example: "Rp %.0f" or "$%.2f". Defaults to "%v" if empty.
	JSONValueFormat string `json:"json_value_format,omitempty"`

	// --- Fields used when Type == "html" ---
	// ItemSelector is the CSS selector matching each repeating item on the page.
	// Example: ".product-card" — every product card on a listing page.
	ItemSelector string `json:"item_selector,omitempty"`

	// TitleSelector is the CSS selector for the title, relative to ItemSelector.
	TitleSelector string `json:"title_selector,omitempty"`

	// ValueSelector is the CSS selector for the value/price, relative to ItemSelector.
	ValueSelector string `json:"value_selector,omitempty"`

	// ValueAttr, if set, extracts an HTML attribute (e.g. "href") from the
	// element matched by ValueSelector instead of its text content.
	// Useful for grabbing a link URL alongside a title. Leave empty to use
	// the element's text as before.
	ValueAttr string `json:"value_attr,omitempty"`

	// MaxItems limits how many matched HTML items to keep (0 = no limit).
	MaxItems int `json:"max_items,omitempty"`
}

// Config is the top-level structure of configs/targets.json.
type Config struct {
	Targets []Target `json:"targets"`
}

// LoadConfig reads and parses the target config file.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return &cfg, nil
}

// EnabledTargets returns only the targets with Enabled == true.
func (c *Config) EnabledTargets() []Target {
	var out []Target
	for _, t := range c.Targets {
		if t.Enabled {
			out = append(out, t)
		}
	}
	return out
}
