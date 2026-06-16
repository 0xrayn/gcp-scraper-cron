# gcp-scraper-cron

A scheduled web scraper that runs daily on Google Cloud Platform using Cloud Functions and Cloud Scheduler. Results are delivered to Telegram automatically — no server required, no laptop needed.

---

## What it does

The scraper wakes up on a cron schedule, pulls data from the web (exchange rates, prices, news, stock availability — whatever you point it at), formats the results, and sends them to a Telegram bot. When it's done, it shuts down. You pay nothing while it's idle.

Out of the box it scrapes USD/IDR/EUR/SGD exchange rates from a free public API. Swap in any source you want by editing one file.

---

## Screenshots

**GCP Cloud Functions dashboard**
![Cloud Functions Dashboard](docs/screenshots/cloud-functions.png)

**Cloud Scheduler job**
![Cloud Scheduler](docs/screenshots/cloud-scheduler.png)

**Telegram notification result**
![Telegram Bot Output](docs/screenshots/telegram-output.png)

**GitHub Actions deploy log**
![GitHub Actions](docs/screenshots/github-actions.png)

> Add your own screenshots to `docs/screenshots/` after deploying.

---

## Architecture

```
GitHub Repository
      │
      │  push to main
      ▼
GitHub Actions
      │
      │  gcloud deploy
      ▼
Cloud Functions Gen2 (Go 1.21)      ◄──── Cloud Scheduler (daily cron)
      │
      ├── internal/scraper      fetch data from web/API
      └── internal/notifier     send results to Telegram
                                        │
                                        ▼
                                  Your Telegram
```

---

## Cost

| Service | Usage | Free tier | Cost |
|---|---|---|---|
| Cloud Functions Gen2 | 30 calls/month | 2M calls/month | free |
| Cloud Scheduler | 1 job | 3 jobs free | free |
| Egress network | < 1 GB/month | 1 GB/month | free |
| Cloud Build | ~5 min/deploy | 120 min/day | free |
| **Total** | | | **$0/month** |

Set a billing alert at $5 in GCP Console just in case.

---

## Project structure

```
gcp-scraper-cron/
├── function.go                    entry point for Cloud Functions
├── go.mod
├── internal/
│   ├── scraper/
│   │   └── scraper.go             scraping logic — edit this to change data source
│   └── notifier/
│       └── telegram.go            formats and sends Telegram messages
├── cmd/
│   └── deploy/
│       └── main.go                local test runner (no deploy needed)
├── docs/
│   └── screenshots/               add your screenshots here
└── .github/
    └── workflows/
        └── deploy.yml             CI/CD pipeline
```

---

## Setup

### Prerequisites

- Google Cloud account (free tier works, comes with $300 credit)
- GitHub account
- [gcloud CLI](https://cloud.google.com/sdk/docs/install) installed
- Go 1.21+
- Telegram account

---

### 1. Create a GCP project

```bash
gcloud auth login
gcloud projects create your-project-id
gcloud config set project your-project-id

gcloud services enable cloudfunctions.googleapis.com
gcloud services enable cloudscheduler.googleapis.com
gcloud services enable cloudbuild.googleapis.com
gcloud services enable run.googleapis.com
```

---

### 2. Create a service account for GitHub Actions

```bash
gcloud iam service-accounts create github-deployer \
  --display-name="GitHub Actions Deployer"

gcloud projects add-iam-policy-binding your-project-id \
  --member="serviceAccount:github-deployer@your-project-id.iam.gserviceaccount.com" \
  --role="roles/cloudfunctions.developer"

gcloud projects add-iam-policy-binding your-project-id \
  --member="serviceAccount:github-deployer@your-project-id.iam.gserviceaccount.com" \
  --role="roles/cloudscheduler.admin"

gcloud projects add-iam-policy-binding your-project-id \
  --member="serviceAccount:github-deployer@your-project-id.iam.gserviceaccount.com" \
  --role="roles/run.admin"

gcloud projects add-iam-policy-binding your-project-id \
  --member="serviceAccount:github-deployer@your-project-id.iam.gserviceaccount.com" \
  --role="roles/iam.serviceAccountUser"

# Export the key — you'll paste this into GitHub Secrets
gcloud iam service-accounts keys create key.json \
  --iam-account=github-deployer@your-project-id.iam.gserviceaccount.com
```

---

### 3. Create a Telegram bot

1. Open Telegram and search for **@BotFather**
2. Send `/newbot` and follow the prompts
3. Copy the **token** you receive
4. Send any message to your new bot
5. Open this URL in a browser to get your chat ID:
   ```
   https://api.telegram.org/bot<YOUR_TOKEN>/getUpdates
   ```
6. Find `"chat":{"id": 123456789}` — that number is your chat ID

---

### 4. Add GitHub Secrets

Go to your repo on GitHub: **Settings → Secrets and variables → Actions → New repository secret**

| Secret | Value |
|---|---|
| `GCP_SA_KEY` | Full contents of `key.json` |
| `GCP_PROJECT_ID` | `your-project-id` |
| `TELEGRAM_TOKEN` | Token from BotFather |
| `TELEGRAM_CHAT_ID` | Your chat ID number |
| `SCHEDULER_SECRET` | Any random string, e.g. `s3cr3t-abc` |

---

### 5. Push and deploy

```bash
git init
git add .
git commit -m "initial commit"
git branch -M main
git remote add origin https://github.com/your-username/gcp-scraper-cron.git
git push -u origin main
```

GitHub Actions deploys automatically. Check the **Actions** tab for progress.

---

## Running locally

Install dependencies first:

```bash
go mod tidy
```

Copy `.env.example` to `.env` and fill in your values:

```
TELEGRAM_TOKEN=token_from_botfather
TELEGRAM_CHAT_ID=your_chat_id
SCHEDULER_SECRET=any_random_string
```

Then run:

```bash
go run ./cmd/deploy
```

The app reads `.env` automatically. The `.env` file is listed in `.gitignore` so it will never be committed to GitHub.

---

**If you prefer setting env variables manually instead:**

**Windows (CMD):**
```cmd
set TELEGRAM_TOKEN=your-token
set TELEGRAM_CHAT_ID=your-chat-id
go run ./cmd/deploy
```

**Windows (PowerShell):**
```powershell
$env:TELEGRAM_TOKEN="your-token"
$env:TELEGRAM_CHAT_ID="your-chat-id"
go run ./cmd/deploy
```

**Mac/Linux:**
```bash
export TELEGRAM_TOKEN=your-token
export TELEGRAM_CHAT_ID=your-chat-id
go run ./cmd/deploy
```

---

## Adding a new scraper

Open `internal/scraper/scraper.go` and add a function:

```go
func scrapeProductPrice(ctx context.Context) ([]Item, error) {
    resp, err := http.Get("https://example.com/product/123")
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    // parse the response here

    return []Item{
        {
            Title:     "Product Name",
            Value:     "Rp 1.200.000",
            Source:    "example.com",
            ScrapedAt: time.Now(),
        },
    }, nil
}
```

Then register it in `Run()`:

```go
func Run(ctx context.Context) ([]Item, error) {
    var results []Item

    currencyItems, _ := scrapeCurrency(ctx)
    results = append(results, currencyItems...)

    priceItems, err := scrapeProductPrice(ctx)
    if err != nil {
        log.Printf("price scraper failed: %v", err) // log and continue, don't stop everything
    } else {
        results = append(results, priceItems...)
    }

    return results, nil
}
```

---

## Cron schedule

Edit `--schedule` in `.github/workflows/deploy.yml`:

| Schedule | Runs |
|---|---|
| `0 0 * * *` | Daily at 07:00 WIB (00:00 UTC) |
| `0 22 * * *` | Daily at 05:00 WIB (22:00 UTC) |
| `0 */6 * * *` | Every 6 hours |
| `0 0 * * 1` | Every Monday |
| `*/30 * * * *` | Every 30 minutes |

Cloud Scheduler uses UTC. Indonesia WIB is UTC+7, so subtract 7 hours from your target time.

---

## Viewing logs

```bash
gcloud functions logs read scraper-harian \
  --region=asia-southeast2 \
  --limit=50
```

---

## Common issues

**`go.sum` missing entries on first run**

```bash
go mod tidy
```

**Deploy fails in GitHub Actions**

Check that all 5 secrets are set correctly. Open the failed workflow run in the Actions tab and read the error output.

**No Telegram message received**

Make sure you sent at least one message to the bot before testing — bots cannot initiate conversations. Verify your token and chat ID are correct:

```bash
curl https://api.telegram.org/bot<TOKEN>/getMe
```

**Function timeout**

Default is 300 seconds. For heavy scraping jobs, add `--timeout=540s` to the deploy command (Gen2 max is 9 minutes).

---

## Tech stack

| | |
|---|---|
| Language | Go 1.21 |
| Compute | GCP Cloud Functions Gen2 |
| Scheduler | GCP Cloud Scheduler |
| CI/CD | GitHub Actions |
| Notifications | Telegram Bot API |
| Framework | functions-framework-go v1.8 |

---

## License

MIT
