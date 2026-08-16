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

func FetchAllGlobalEvents(ctx context.Context) ([]models.TimelineEvent, error) {
	fetchTime := time.Now().UnixMilli()

	client := &http.Client{
		Timeout: 8 * time.Second,
	}

	appToken := os.Getenv("APP_TOKEN")
	if appToken == "" {
		appToken = "c16198fe11e3b04c8f5f403487c672b1"
	}

	// 1. Fetch CricHD channel stream lookup map
	var crichdStreamMap map[string]models.CrichdApiStream
	req, err := http.NewRequestWithContext(ctx, "GET", "https://raw.githubusercontent.com/abusaeeidx/CricHd-playlists-Auto-Update-permanent/main/api.json", nil)
	if err == nil {
		req.Header.Set("X-App-Token", appToken)
		resp, err := client.Do(req)
		if err == nil && resp.StatusCode == 200 {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			crichdStreamMap = ParseCrichdStreamMap(body)
		}
	}

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

			reqCtx, cancel := context.WithTimeout(ctx, 6*time.Second)
			defer cancel()

			req, err := http.NewRequestWithContext(reqCtx, "GET", sc.URL, nil)
			if err != nil {
				return
			}
			req.Header.Set("X-App-Token", appToken)

			resp, err := client.Do(req)
			if err != nil || resp.StatusCode != 200 {
				if resp != nil {
					resp.Body.Close()
				}
				log.Printf("Failed to fetch source %s: %v", sc.URL, err)
				return
			}
			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
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

	// 5. Filter out ended events & streamless live events
	now := time.Now().UnixMilli()
	var finalEvents []models.TimelineEvent

	for _, ev := range unifiedEvents {
		isLiveNow := (ev.IsLive != nil && *ev.IsLive) || (now >= ev.StartTimestampMs && now <= ev.EndTimestampMs)

		// Filter events that ended more than 1 hour ago
		if !isLiveNow && ev.EndTimestampMs <= (now-60*60*1000) {
			continue
		}

		// Filter live events with no valid playable streams
		if isLiveNow && (len(ev.Streams) == 0) {
			continue
		}

		finalEvents = append(finalEvents, ev)
	}

	// 6. Sort chronologically by start time
	sort.Slice(finalEvents, func(i, j int) bool {
		return finalEvents[i].StartTimestampMs < finalEvents[j].StartTimestampMs
	})

	return finalEvents, nil
}
