// Package scraper berisi logika utama untuk scraping data.
// Contoh ini scrape harga mata uang USD/IDR dari API publik.
// Bisa diganti dengan scraper web biasa (berita, produk, dll).
package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Item mewakili satu hasil scrape
type Item struct {
	Title     string    `json:"title"`
	Value     string    `json:"value"`
	Source    string    `json:"source"`
	ScrapedAt time.Time `json:"scraped_at"`
}

// Run menjalankan semua scraper dan menggabungkan hasilnya
func Run(ctx context.Context) ([]Item, error) {
	var results []Item

	// Scraper 1: Kurs mata uang (gratis, tidak butuh API key)
	currencyItems, err := scrapeCurrency(ctx)
	if err != nil {
		return nil, fmt.Errorf("currency scraper: %w", err)
	}
	results = append(results, currencyItems...)

	// Scraper 2: Tambahkan scraper lain di sini
	// newsItems, err := scrapeNews(ctx)
	// results = append(results, newsItems...)

	return results, nil
}

// scrapeCurrency mengambil kurs USD/IDR dari API publik frankfurter.app
func scrapeCurrency(ctx context.Context) ([]Item, error) {
	url := "https://api.frankfurter.app/latest?from=USD&to=IDR,EUR,SGD"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request gagal: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status tidak OK: %d", resp.StatusCode)
	}

	// Parse response JSON
	var data struct {
		Base  string             `json:"base"`
		Date  string             `json:"date"`
		Rates map[string]float64 `json:"rates"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("parse JSON gagal: %w", err)
	}

	// Konversi ke format Item
	now := time.Now()
	items := []Item{
		{
			Title:     "USD → IDR",
			Value:     fmt.Sprintf("Rp %.0f", data.Rates["IDR"]),
			Source:    "frankfurter.app",
			ScrapedAt: now,
		},
		{
			Title:     "USD → EUR",
			Value:     fmt.Sprintf("€ %.4f", data.Rates["EUR"]),
			Source:    "frankfurter.app",
			ScrapedAt: now,
		},
		{
			Title:     "USD → SGD",
			Value:     fmt.Sprintf("S$ %.4f", data.Rates["SGD"]),
			Source:    "frankfurter.app",
			ScrapedAt: now,
		},
	}

	return items, nil
}

// Contoh scraper web HTML menggunakan net/http + string parsing sederhana.
// Untuk scraping HTML lebih kompleks, gunakan library "golang.org/x/net/html"
// atau "github.com/PuerkitoBio/goquery"
//
// func scrapeNews(ctx context.Context) ([]Item, error) {
//     resp, err := http.Get("https://example.com/news")
//     if err != nil { return nil, err }
//     defer resp.Body.Close()
//
//     doc, err := goquery.NewDocumentFromReader(resp.Body)
//     if err != nil { return nil, err }
//
//     var items []Item
//     doc.Find("article.news-item").Each(func(i int, s *goquery.Selection) {
//         items = append(items, Item{
//             Title:     s.Find("h2").Text(),
//             Value:     s.Find("p.summary").Text(),
//             Source:    "example.com",
//             ScrapedAt: time.Now(),
//         })
//     })
//     return items, nil
// }
