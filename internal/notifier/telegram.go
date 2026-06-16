// Package notifier mengirimkan hasil scrape ke Telegram.
// Untuk menggunakan, set environment variable:
//   TELEGRAM_TOKEN  = token dari @BotFather
//   TELEGRAM_CHAT_ID = chat ID tujuan (bisa group atau personal)
package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/yourusername/gcp-scraper-cron/internal/scraper"
)

// Send mengirim semua hasil scrape ke Telegram
func Send(ctx context.Context, items []scraper.Item) error {
	token := os.Getenv("TELEGRAM_TOKEN")
	chatID := os.Getenv("TELEGRAM_CHAT_ID")

	if token == "" || chatID == "" {
		// Jika tidak ada konfigurasi Telegram, skip (tidak error)
		return nil
	}

	message := formatMessage(items)
	return sendTelegram(ctx, token, chatID, message)
}

// formatMessage membuat pesan Telegram yang rapi dengan Markdown
func formatMessage(items []scraper.Item) string {
	var sb strings.Builder

	sb.WriteString("📊 *Laporan Harian Scraper*\n")
	sb.WriteString(fmt.Sprintf("🕐 %s WIB\n\n", time.Now().Format("02 Jan 2006, 15:04")))

	// Kelompokkan per source
	sourceMap := make(map[string][]scraper.Item)
	for _, item := range items {
		sourceMap[item.Source] = append(sourceMap[item.Source], item)
	}

	for source, sourceItems := range sourceMap {
		sb.WriteString(fmt.Sprintf("*%s*\n", source))
		for _, item := range sourceItems {
			sb.WriteString(fmt.Sprintf("• %s: `%s`\n", item.Title, item.Value))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("_Total: %d item_", len(items)))
	return sb.String()
}

// sendTelegram melakukan HTTP POST ke Telegram Bot API
func sendTelegram(ctx context.Context, token, chatID, message string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)

	payload, err := json.Marshal(map[string]string{
		"chat_id":    chatID,
		"text":       message,
		"parse_mode": "Markdown",
	})
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("kirim pesan: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API error: status %d", resp.StatusCode)
	}

	return nil
}
