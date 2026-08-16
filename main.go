package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"strings"
	"time"

	"github.com/neelniloy/lstv-event-parser/parser"
	"github.com/neelniloy/lstv-event-parser/uploader"
)

func loadDotEnv(filepath string) {
	file, err := os.Open(filepath)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			if os.Getenv(key) == "" {
				os.Setenv(key, val)
			}
		}
	}
}

func main() {
	loadDotEnv(".env")

	dryRunFlag := flag.Bool("dry-run", false, "Save output locally to events.json instead of uploading to Cloudflare R2")
	outputFileFlag := flag.String("output", "events.json", "Local output filename for dry-run")
	flag.Parse()

	startTime := time.Now()
	log.Println("Starting Live Sports Events parser...")

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	events, err := parser.FetchAllGlobalEvents(ctx)
	if err != nil {
		log.Fatalf("Error parsing events: %v", err)
	}

	log.Printf("Successfully fetched and normalized %d sports events in %v", len(events), time.Since(startTime))

	jsonData, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		log.Fatalf("Error marshaling events JSON: %v", err)
	}

	dryRunEnv := os.Getenv("DRY_RUN") == "true" || *dryRunFlag

	if dryRunEnv {
		err := os.WriteFile(*outputFileFlag, jsonData, 0644)
		if err != nil {
			log.Fatalf("Failed to write dry-run file: %v", err)
		}
		log.Printf("Dry-run complete. Saved %d events to %s", len(events), *outputFileFlag)
		return
	}

	getEnv := func(keys ...string) string {
		for _, key := range keys {
			if val := os.Getenv(key); val != "" {
				return val
			}
		}
		return ""
	}

	r2Cfg := uploader.R2Config{
		AccountID:       getEnv("CF_R2_ACCOUNT_ID", "R2_ACCOUNT_ID"),
		AccessKeyID:     getEnv("CF_R2_ACCESS_KEY_ID", "R2_ACCESS_KEY_ID"),
		SecretAccessKey: getEnv("CF_R2_SECRET_ACCESS_KEY", "R2_SECRET_ACCESS_KEY"),
		BucketName:      getEnv("CF_R2_BUCKET_NAME", "R2_BUCKET_NAME"),
	}

	if r2Cfg.AccountID == "" || r2Cfg.AccessKeyID == "" || r2Cfg.SecretAccessKey == "" || r2Cfg.BucketName == "" ||
		r2Cfg.AccountID == "your_cloudflare_account_id" {
		log.Println("No valid Cloudflare R2 credentials found in .env or environment. Saving locally to events.json...")
		_ = os.WriteFile("events.json", jsonData, 0644)
		return
	}

	err = uploader.UploadToR2(ctx, r2Cfg, "events.json", jsonData)
	if err != nil {
		log.Fatalf("Failed to upload to Cloudflare R2: %v", err)
	}

	log.Printf("All tasks completed successfully in %v!", time.Since(startTime))
}

