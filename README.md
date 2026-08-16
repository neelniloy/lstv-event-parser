# ⚡ LSTV Live Events Parser

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![Automated Sync](https://github.com/neelniloy/lstv-event-parser/actions/workflows/parse.yml/badge.svg)](https://github.com/neelniloy/lstv-event-parser/actions/workflows/parse.yml)
[![Storage](https://img.shields.io/badge/Storage-Cloudflare_R2-F38020?style=flat&logo=cloudflare)](https://www.cloudflare.com/developer-platform/r2/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

A high-performance **Go microservice** that aggregates live sports events from 10+ external sources, normalizes event metadata, deduplicates & merges multi-server stream links, dynamically injects domain headers/DRM configs, and automatically deploys a single consolidated `events.json` payload to Cloudflare R2 every 5 minutes.

---

## 🏗 System Architecture

```
┌─────────────────────────────────────────┐
│     10+ External Sports Data Sources     │
│   (Tapmad, SonyLiv, CricHD, Willow, etc.)│
└────────────────────┬────────────────────┘
                     │ Concurrent Fetching (Goroutines)
                     ▼
┌─────────────────────────────────────────┐
│       ⚡ LSTV Event Parser Service       │
│  - Match Key Deduplication              │
│  - Stream Link Aggregation              │
│  - Header & DRM Injection               │
│  - Chronological Sorting & Filtering    │
└────────────────────┬────────────────────┘
                     │ S3 Upload (~1 sec)
                     ▼
┌─────────────────────────────────────────┐
│           Cloudflare R2 Bucket          │
│          (CDN Egress: $0 / Mo)          │
└────────────────────┬────────────────────┘
                     │ Single GET Request
                     ▼
┌─────────────────────────────────────────┐
│   Client Apps (Android / Fire TV / Web) │
└─────────────────────────────────────────┘
```

---

## 🚀 Features

- **⚡ Blazing Fast Concurrency**: Fetches and parses 300+ events across 10 JSON endpoints in **~1 second** using Go goroutines and channels.
- **🔄 Smart Match Deduplication**: Normalizes team names (`barcelona_vs_realmadrid`) to automatically merge stream servers across different providers into a single unified event object.
- **🛡 Domain-Aware Header & DRM Injection**: Automatically injects required HTTP headers (`Referer`, `Origin`, `User-Agent`) and DRM configurations (`ClearKey`) tailored per stream host.
- **📊 Sport Categorization**: Intelligent classification engine targeting Cricket, Football, Motorsport, Combat Sports, Tennis, Basketball, and more.
- **☁️ Zero-Maintenance Serverless Sync**: Deploys `events.json` to Cloudflare R2 every 5 minutes via scheduled GitHub Actions, ensuring clients query a static CDN edge.

---

## 🛠 Local Setup & Running

### Prerequisites
- [Go 1.21+](https://golang.org/dl/) installed.

### 1. Clone & Install
```bash
git clone https://github.com/neelniloy/lstv-event-parser.git
cd lstv-event-parser
go mod download
```

### 2. Local Dry-Run (No Upload)
To test parsing and generate `events.json` locally without uploading to R2:
```bash
go run main.go -dry-run
```

### 3. Full Run with Local `.env`
Copy `.env.example` to `.env` and fill in your Cloudflare R2 credentials:
```bash
cp .env.example .env
go run main.go
```

---

## 🔑 GitHub Actions Setup (Automated Deployment)

To configure automatic sync every 5 minutes via GitHub Actions:

1. Create a public Cloudflare R2 Bucket (e.g. `lstv`).
2. Generate R2 API credentials in your Cloudflare Dashboard.
3. In your GitHub repository, navigate to **Settings > Secrets and variables > Actions** and add the following 4 secrets:

| Secret Name | Description / Example |
| :--- | :--- |
| `CF_R2_ACCOUNT_ID` | Cloudflare Account ID |
| `CF_R2_ACCESS_KEY_ID` | Cloudflare R2 Access Key ID |
| `CF_R2_SECRET_ACCESS_KEY` | Cloudflare R2 Secret Access Key |
| `CF_R2_BUCKET_NAME` | `lstv` |

*(Note: `R2_ACCOUNT_ID`, `R2_ACCESS_KEY_ID`, `R2_SECRET_ACCESS_KEY`, and `R2_BUCKET_NAME` are also supported as fallbacks).*

---

## 📄 Output Data Schema Example (`events.json`)

```json
[
  {
    "id": "realmadrid_vs_barcelona",
    "title": "Real Madrid vs Barcelona",
    "category": "Football",
    "league": "La Liga",
    "homeTeam": "Real Madrid",
    "awayTeam": "Barcelona",
    "startTimestampMs": 1723820400000,
    "endTimestampMs": 1723831200000,
    "isLive": true,
    "streams": [
      {
        "sourceName": "SonyLiv",
        "streamUrl": "https://example.com/live/stream1/index.m3u8",
        "type": "hls",
        "isM3u": true,
        "headers": {
          "Origin": "https://www.sonyliv.com",
          "Referer": "https://www.sonyliv.com/",
          "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64)..."
        }
      }
    ]
  }
]
```

---

## 📄 License

Distributed under the MIT License. See `LICENSE` for more information.

