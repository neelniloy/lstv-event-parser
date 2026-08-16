# ⚡ LSTV Live Events Parser (Go Microservice)

High-performance Go service that fetches live sports events from 10+ external sources, normalizes event structures, merges duplicate streams across multiple server links, and automatically deploys a single consolidated `events.json` to Cloudflare R2 every 5 minutes via GitHub Actions.

---

## 🚀 Features

- **Blazing Fast**: Runs parallel goroutines across 10 JSON endpoints in ~1–2 seconds.
- **Smart Merging**: Normalizes match keys (`barcelona_vs_realmadrid`) and combines server stream links.
- **Robust Headers & DRM**: Auto-injects necessary HTTP `Referer`, `Origin`, and `User-Agent` domain headers.
- **Auto Classification**: Categorizes events into Cricket, Football, Motorsport, Combat Sports, Tennis, Basketball, etc.
- **Cloudflare R2 Direct Sync**: Uploads `events.json` using AWS S3 v2 SDK.

---

## 🛠 Local Setup & Dry Run

To test parsing locally without uploading to Cloudflare R2:

```bash
# Clone the repository
git clone https://github.com/neelniloy/lstv-event-parser.git
cd lstv-event-parser

# Run in dry-run mode (generates events.json locally)
go run main.go -dry-run
```

---

## 🔑 GitHub Actions Setup & Cloudflare R2 Secrets

1. Create a public Cloudflare R2 Bucket (e.g. `lstv-events`).
2. Generate an API Token in Cloudflare R2 dashboard with **Admin / Object Read & Write** permissions.
3. In your GitHub repository settings, navigate to **Settings > Secrets and variables > Actions** and add the following 4 secrets:

| Secret Name | Description / Example Value |
|---|---|
| `R2_ACCOUNT_ID` | Cloudflare Account ID (Found in R2 dashboard URL / right sidebar) |
| `R2_ACCESS_KEY_ID` | Cloudflare R2 Access Key ID |
| `R2_SECRET_ACCESS_KEY` | Cloudflare R2 Secret Access Key |
| `R2_BUCKET_NAME` | `lstv-events` (Your R2 bucket name) |

---

## 🌐 Fetching from Client App (Android / TV)

Once uploaded to R2, your Android client app only makes a single GET request:

```http
GET https://pub-<hash>.r2.dev/events.json
```
or via custom domain:
```http
GET https://events.yourdomain.com/events.json
```
