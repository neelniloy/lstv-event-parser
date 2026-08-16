package parser

import (
	"fmt"

	"math"
	"regexp"
	"strings"

	"github.com/neelniloy/lstv-event-parser/models"
)

var nonAlphanumericRegex = regexp.MustCompile(`[^a-z0-9]`)

var teamAliases = map[string]string{
	// Football
	"manutd":            "manchesterunited",
	"mancity":           "manchestercity",
	"barca":             "barcelona",
	"fcbarcelona":       "barcelona",
	"realmadridcf":      "realmadrid",
	"atleticomadrid":    "atletico",
	"parissaintgermain": "psg",
	"parissg":           "psg",
	"juve":              "juventus",
	"inter":             "intermilan",
	"acmilan":           "milan",
	"spurs":             "tottenham",
	"tottenhamhotspur":  "tottenham",
	"bayernmunich":      "bayern",
	"bayernmunchen":     "bayern",
	"bvb":               "borussiadortmund",
	"dortmund":          "borussiadortmund",
	"celtadevigo":       "celta",
	"celtavigo":         "celta",
	"villarrealcf":      "villarreal",
	"athleticbilbao":    "athletic",
	"athleticclub":      "athletic",
	"westhamunited":     "westham",
	"wolverhamptonwanderers": "wolverhampton",
	"wolves":           "wolverhampton",
	"nottinghamforest":  "nottingham",
	"sheffieldunited":   "sheffieldutd",

	// Baseball / US Sports
	"texasrangers": "rangers",

	// International Countries
	"ind": "india",
	"eng": "england",
	"aus": "australia",
	"pak": "pakistan",
	"ban": "bangladesh",
	"bd":  "bangladesh",
	"nz":  "newzealand",
	"sa":  "southafrica",
	"wi":  "westindies",
	"sl":  "srilanka",
	"afg": "afghanistan",
	"ned": "netherlands",
	"ire": "ireland",
	"zw":  "zimbabwe",
	"zim": "zimbabwe",

	// Cricket Leagues
	"rcb":  "royalchallengersbengaluru",
	"csk":  "chennaisuperkings",
	"mi":   "mumbaiindians",
	"kkr":  "kolkataknightriders",
	"dc":   "delhicapitals",
	"pkbs": "punjabkings",
	"pbks": "punjabkings",
	"rr":   "rajasthanroyals",
	"srh":  "sunrisershyderabad",
	"gt":   "gujarattitans",
	"lsg":  "lucknowsupergiants",
}

var titleNoiseWords = []string{
	"1st test", "2nd test", "3rd test", "4th test", "5th test",
	"1st t20", "2nd t20", "3rd t20", "4th t20", "5th t20",
	"1st odi", "2nd odi", "3rd odi", "4th odi", "5th odi",
	"t20i", "t20", "odi", "test series", "test match",
	"match 1", "match 2", "match 3", "match 4", "match 5",
	"game 1", "game 2", "game 3",
	"day 1", "day 2", "day 3", "day 4", "day 5",
	"live", "hd", "fhd", "4k", "stream", "server 1", "server 2",
}

func CleanTitleNoise(title string) string {
	t := strings.ToLower(title)
	for _, nw := range titleNoiseWords {
		t = strings.ReplaceAll(t, nw, "")
	}
	return strings.TrimSpace(t)
}

func NormalizeTeamName(name string) string {
	cleaned := nonAlphanumericRegex.ReplaceAllString(strings.ToLower(name), "")
	if alias, found := teamAliases[cleaned]; found {
		return alias
	}
	return cleaned
}

func GetMatchKey(event *models.TimelineEvent) string {
	home := ""
	away := ""
	if event.HomeTeam != "" {
		home = NormalizeTeamName(event.HomeTeam)
	}
	if event.AwayTeam != "" {
		away = NormalizeTeamName(event.AwayTeam)
	}

	if home != "" && away != "" && home != away && home != "women" && away != "women" {
		if home < away {
			return fmt.Sprintf("%s_vs_%s", home, away)
		}
		return fmt.Sprintf("%s_vs_%s", away, home)
	}

	titleLower := CleanTitleNoise(event.Title)
	splitters := []string{" vs ", " vs. ", " v ", " @ ", " - "}
	for _, spl := range splitters {
		if strings.Contains(titleLower, spl) {
			parts := strings.Split(titleLower, spl)
			if len(parts) >= 2 {
				t1 := NormalizeTeamName(parts[0])
				t2 := NormalizeTeamName(parts[1])
				if t1 != "" && t2 != "" {
					if t1 < t2 {
						return fmt.Sprintf("%s_vs_%s", t1, t2)
					}
					return fmt.Sprintf("%s_vs_%s", t2, t1)
				}
			}
		}
	}

	return NormalizeTeamName(titleLower)
}

func IsMatchingEvent(existing, event *models.TimelineEvent) bool {
	timeDiff := math.Abs(float64(existing.StartTimestampMs - event.StartTimestampMs))
	if timeDiff > 4.0*60.0*60.0*1000.0 {
		return false
	}

	key1 := GetMatchKey(existing)
	key2 := GetMatchKey(event)

	// 1. Exact key match
	if key1 == key2 && key1 != "" {
		return true
	}

	// 2. Substring key match (e.g. australia_vs_bangladesh vs 1sttestaustralia_vs_bangladesh)
	if len(key1) >= 5 && len(key2) >= 5 {
		if strings.Contains(key1, key2) || strings.Contains(key2, key1) {
			return true
		}
	}

	// 3. Team name fuzzy match
	h1, a1 := NormalizeTeamName(existing.HomeTeam), NormalizeTeamName(existing.AwayTeam)
	h2, a2 := NormalizeTeamName(event.HomeTeam), NormalizeTeamName(event.AwayTeam)

	if h1 != "" && a1 != "" && h2 != "" && a2 != "" {
		sameOrder := (strings.Contains(h1, h2) || strings.Contains(h2, h1)) &&
			(strings.Contains(a1, a2) || strings.Contains(a2, a1))
		revOrder := (strings.Contains(h1, a2) || strings.Contains(a2, h1)) &&
			(strings.Contains(a1, h2) || strings.Contains(h2, a1))

		if sameOrder || revOrder {
			return true
		}
	}

	return false
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
		matchIndex := -1

		for i, existing := range mergedList {
			if IsMatchingEvent(&existing, &event) {
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
