package parser

import (
	"fmt"

	"math"
	"regexp"
	"strings"

	"github.com/neelniloy/lstv-event-parser/models"
)

var nonAlphanumericRegex = regexp.MustCompile(`[^a-z0-9]`)

func GetMatchKey(event *models.TimelineEvent) string {
	home := ""
	away := ""
	if event.HomeTeam != "" {
		home = nonAlphanumericRegex.ReplaceAllString(strings.ToLower(event.HomeTeam), "")
	}
	if event.AwayTeam != "" {
		away = nonAlphanumericRegex.ReplaceAllString(strings.ToLower(event.AwayTeam), "")
	}

	if home != "" && away != "" && home != away && home != "women" && away != "women" {
		if home < away {
			return fmt.Sprintf("%s_vs_%s", home, away)
		}
		return fmt.Sprintf("%s_vs_%s", away, home)
	}

	titleLower := strings.ToLower(event.Title)
	splitters := []string{" vs ", " vs. ", " v ", " @ ", " - "}
	for _, spl := range splitters {
		if strings.Contains(titleLower, spl) {
			parts := strings.Split(titleLower, spl)
			if len(parts) >= 2 {
				t1 := strings.TrimSpace(nonAlphanumericRegex.ReplaceAllString(parts[0], ""))
				t2 := strings.TrimSpace(nonAlphanumericRegex.ReplaceAllString(parts[1], ""))
				if t1 != "" && t2 != "" {
					if t1 < t2 {
						return fmt.Sprintf("%s_vs_%s", t1, t2)
					}
					return fmt.Sprintf("%s_vs_%s", t2, t1)
				}
			}
		}
	}

	return nonAlphanumericRegex.ReplaceAllString(titleLower, "")
}

func IsDynamicLiveTime(eventTime int64, fetchTime int64) bool {
	return math.Abs(float64(eventTime-fetchTime)) < 30000.0
}

func MergeAndDeduplicate(events []models.TimelineEvent, fetchTime int64) []models.TimelineEvent {
	if len(events) == 0 {
		return nil
	}

	mergedList := make([]models.TimelineEvent, 0, len(events))

	for _, event := range events {
		key := GetMatchKey(&event)
		matchIndex := -1

		for i, existing := range mergedList {
			existingKey := GetMatchKey(&existing)
			timeDiff := math.Abs(float64(existing.StartTimestampMs - event.StartTimestampMs))
			if existingKey == key && timeDiff <= 4.0*60.0*60.0*1000.0 {
				matchIndex = i
				break
			}
		}

		if matchIndex != -1 {
			existing := mergedList[matchIndex]

			// Combine stream servers
			combinedStreams := make([]models.EventStream, 0, len(existing.Streams)+len(event.Streams))
			combinedStreams = append(combinedStreams, existing.Streams...)

			for _, ns := range event.Streams {
				exists := false
				for _, cs := range combinedStreams {
					if cs.StreamURL == ns.StreamURL {
						exists = true
						break
					}
				}
				if !exists {
					combinedStreams = append(combinedStreams, ns)
				}
			}

			existingIsDynamic := existing.IsStartTimeDynamic || IsDynamicLiveTime(existing.StartTimestampMs, fetchTime)
			newIsDynamic := event.IsStartTimeDynamic || IsDynamicLiveTime(event.StartTimestampMs, fetchTime)

			var chosenStart int64
			var chosenEnd int64

			switch {
			case existingIsDynamic && !newIsDynamic:
				chosenStart = event.StartTimestampMs
				chosenEnd = event.EndTimestampMs
			case !existingIsDynamic && newIsDynamic:
				chosenStart = existing.StartTimestampMs
				chosenEnd = existing.EndTimestampMs
			default:
				diff := math.Abs(float64(existing.StartTimestampMs - event.StartTimestampMs))
				if diff <= 90.0*60.0*1000.0 {
					if existing.StartTimestampMs > event.StartTimestampMs {
						chosenStart = existing.StartTimestampMs
						chosenEnd = existing.EndTimestampMs
					} else {
						chosenStart = event.StartTimestampMs
						chosenEnd = event.EndTimestampMs
					}
				} else {
					if existing.StartTimestampMs < event.StartTimestampMs {
						chosenStart = existing.StartTimestampMs
						chosenEnd = existing.EndTimestampMs
					} else {
						chosenStart = event.StartTimestampMs
						chosenEnd = event.EndTimestampMs
					}
				}
			}

			merged := existing
			merged.StartTimestampMs = chosenStart
			merged.EndTimestampMs = chosenEnd
			merged.IsStartTimeDynamic = existing.IsStartTimeDynamic && event.IsStartTimeDynamic

			if merged.Description == "" {
				merged.Description = event.Description
			}
			if merged.Category == "" {
				merged.Category = event.Category
			}
			if merged.League == "" {
				merged.League = event.League
			}
			if merged.HomeTeam == "" {
				merged.HomeTeam = event.HomeTeam
			}
			if merged.HomeTeamLogo == "" {
				merged.HomeTeamLogo = event.HomeTeamLogo
			}
			if merged.AwayTeam == "" {
				merged.AwayTeam = event.AwayTeam
			}
			if merged.AwayTeamLogo == "" {
				merged.AwayTeamLogo = event.AwayTeamLogo
			}
			if merged.Poster == "" {
				merged.Poster = event.Poster
			}

			isLiveCombined := false
			if (existing.IsLive != nil && *existing.IsLive) || (event.IsLive != nil && *event.IsLive) {
				isLiveCombined = true
				merged.IsLive = &isLiveCombined
			}

			if len(combinedStreams) > 0 {
				merged.Streams = combinedStreams
			}

			mergedList[matchIndex] = merged
		} else {
			mergedList = append(mergedList, event)
		}
	}

	return mergedList
}
