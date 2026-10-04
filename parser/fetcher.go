package parser

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/neelniloy/lstv-event-parser/models"
)

type SourceConfig struct {
	URL  string
	Type string
}

const defaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

func fetchWithRetry(ctx context.Context, client *http.Client, url string, appToken string, maxRetries int) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt*200) * time.Millisecond)
		}

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("X-App-Token", appToken)
		req.Header.Set("User-Agent", defaultUserAgent)

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			lastErr = http.ErrHandlerTimeout
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}

		return body, nil
	}
	return nil, lastErr
}

func FetchAllGlobalEvents(ctx context.Context) ([]models.TimelineEvent, error) {
	fetchTime := time.Now().UnixMilli()

	client := &http.Client{
		Timeout: 8 * time.Second,
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     30 * time.Second,
		},
	}

	appToken := os.Getenv("APP_TOKEN")
	if appToken == "" {
		appToken = "c16198fe11e3b04c8f5f403487c672b1"
	}

	// 1. Fetch CricHD channel stream lookup map
	var crichdStreamMap map[string]models.CrichdApiStream
	var cricWg sync.WaitGroup
	cricWg.Add(1)

	go func() {
		defer cricWg.Done()
		cricCtx, cancel := context.WithTimeout(ctx, 6*time.Second)
		defer cancel()
		body, err := fetchWithRetry(cricCtx, client, "https://raw.githubusercontent.com/abusaeeidx/CricHd-playlists-Auto-Update-permanent/main/api.json", appToken, 1)
		if err == nil {
			crichdStreamMap = ParseCrichdStreamMap(body)
		}
	}()

	cricWg.Wait()

	// 2. Define the 9 event sources
	sources := []SourceConfig{
		{URL: "https://raw.githubusercontent.com/srhady/tapmad-bd/refs/heads/main/tapmad_bd.json", Type: "tapmad"},
		{URL: "https://raw.githubusercontent.com/srhady/crichd-speical-live-event/refs/heads/main/Live_Events.json", Type: "crichd_live"},
		{URL: "https://raw.githubusercontent.com/srhady/crichd-speical-live-event/refs/heads/main/Footy_Live.json", Type: "crichd_footy"},
		{URL: "https://raw.githubusercontent.com/srhady/bingstream/refs/heads/main/playlist.json", Type: "bingstream"},
		{URL: "https://raw.githubusercontent.com/srhady/axsports/refs/heads/main/live_sports.json", Type: "axsports"},
		{URL: "https://raw.githubusercontent.com/srhady/SonyLiv/refs/heads/main/sonyliv_playlist.json", Type: "sonyliv"},
		{URL: "https://raw.githubusercontent.com/srhady/willow-event/refs/heads/main/live_sports.json", Type: "willow_sports"},
		{URL: "https://raw.githubusercontent.com/srhady/willow-event/refs/heads/main/primevideo_sports.json", Type: "prime_sports"},
		{URL: "https://raw.githubusercontent.com/sm-monirulislam/Upcoming-and-Live-Sports-Data/refs/heads/main/Sports_data.json", Type: "sports_data"},
	}

	var wg sync.WaitGroup
	resultsChan := make(chan []models.TimelineEvent, len(sources))

	for _, src := range sources {
		wg.Add(1)
		go func(sc SourceConfig) {
			defer wg.Done()

			reqCtx, cancel := context.WithTimeout(ctx, 7*time.Second)
			defer cancel()

			body, err := fetchWithRetry(reqCtx, client, sc.URL, appToken, 1)
			if err != nil {
				log.Printf("Failed to fetch source %s: %v", sc.URL, err)
				return
			}

			var parsed []models.TimelineEvent
			switch sc.Type {
			case "tapmad":
				parsed = ParseTapmad(body)
			case "crichd_live":
				parsed = ParseCrichd(body, "Crichd", crichdStreamMap)
			case "crichd_footy":
				parsed = ParseCrichd(body, "Footy", crichdStreamMap)
			case "bingstream":
				parsed = ParseUnifiedMatchesJson(body, "bing")
			case "axsports":
				parsed = ParseUnifiedMatchesJson(body, "ax")
			case "sonyliv":
				parsed = ParseSonyLiv(body)
			case "willow_sports", "prime_sports":
				parsed = ParseWillow(body)
			case "sports_data":
				parsed = ParseSportsData(body)
			}

			resultsChan <- parsed
		}(src)
	}

	wg.Wait()
	close(resultsChan)

	var allEvents []models.TimelineEvent
	for evs := range resultsChan {
		allEvents = append(allEvents, evs...)
	}

	// 3. Assign standardized sport category
	for i := range allEvents {
		allEvents[i].Category = GetSportCategory(&allEvents[i])
	}

	// 4. Merge & Deduplicate
	unifiedEvents := MergeAndDeduplicate(allEvents, fetchTime)

	// 5. Filter out truly backdated events & streamless live events
	now := time.Now().UnixMilli()
	const finishedGracePeriodMs int64 = 20 * 60 * 1000 // 20 minutes grace period after match end
	var finalEvents []models.TimelineEvent

	for _, ev := range unifiedEvents {
		durationMs := GetSportMaxDurationMs(ev.Category, ev.Title)
		maxEndTimeMs := ev.StartTimestampMs + durationMs
		if ev.EndTimestampMs > ev.StartTimestampMs && ev.EndTimestampMs < maxEndTimeMs {
			maxEndTimeMs = ev.EndTimestampMs
		}

		isDynamic := ev.IsStartTimeDynamic
		isUpcoming := !isDynamic && now < ev.StartTimestampMs
		isPastEndTime := !isDynamic && (now > maxEndTimeMs)

		// Determine if the event is currently live:
		// 1) Dynamic streams (24/7 channels or no fixed start time) are live if explicitly marked live.
		// 2) Fixed-time events are live ONLY if now is between start time and realistic end time.
		//    Even if a source feed claims it is still "LIVE", once now > maxEndTimeMs, the match is over!
		isLiveNow := false
		if isDynamic {
			isLiveNow = ev.IsLive != nil && *ev.IsLive
		} else {
			if now >= ev.StartTimestampMs && !isPastEndTime {
				isLiveNow = true
			}
		}

		// Filter out events that ended past the grace period
		if !isDynamic && (now > maxEndTimeMs+finishedGracePeriodMs) {
			continue
		}

		// Filter out stale non-upcoming, non-live events
		if !isUpcoming && !isLiveNow && !isDynamic {
			continue
		}

		// Filter out live events with no valid playable streams
		if isLiveNow && len(ev.Streams) == 0 {
			continue
		}

		// Set calculated IsLive status
		liveFlag := isLiveNow
		ev.IsLive = &liveFlag

		// Ensure EndTimestampMs reflects the realistic end time
		if ev.EndTimestampMs <= ev.StartTimestampMs || ev.EndTimestampMs > maxEndTimeMs {
			ev.EndTimestampMs = maxEndTimeMs
		}

		finalEvents = append(finalEvents, ev)
	}

	// 6. Sort chronologically by start time
	sort.Slice(finalEvents, func(i, j int) bool {
		return finalEvents[i].StartTimestampMs < finalEvents[j].StartTimestampMs
	})

	return finalEvents, nil
}
