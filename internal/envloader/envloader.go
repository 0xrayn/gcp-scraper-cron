// Package envloader provides a minimal .env file loader for local development.
// It is not used in Cloud Functions or Cloud Run — those get env vars injected
// directly by GCP. This is only for `go run` on your own machine.
package envloader

import (
	"bufio"
	"os"
	"strings"
)

// Load reads a .env file and sets any variables not already present
// in the current environment. Missing file is not an error — it just
// means nothing gets loaded, which is fine for production environments
// that set env vars another way.
func Load(filename string) {
	f, err := os.Open(filename)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}
}
