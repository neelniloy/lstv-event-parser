package parser

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"strings"
	"time"

	"github.com/neelniloy/lstv-event-parser/models"
)

const DefaultEventDurationMs int64 = 3 * 60 * 60 * 1000

func HashString(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func ParseCrichdStreamMap(body []byte) map[string]models.CrichdApiStream {
	var items []models.CrichdApiItem
	if err := json.Unmarshal(body, &items); err != nil {
		return nil
	}
	res := make(map[string]models.CrichdApiStream)
	for _, item := range items {
		if strings.TrimSpace(item.Link) == "" {
			continue
		}
		stream := models.CrichdApiStream{
			Link:    item.Link,
			Referer: item.Referer,
			Origin:  item.Origin,
		}
		if idKey := strings.ToLower(strings.TrimSpace(item.ID)); idKey != "" {
			res[idKey] = stream
		}
		if nameKey := nonAlphanumericRegex.ReplaceAllString(strings.ToLower(strings.TrimSpace(item.Name)), ""); nameKey != "" {
			res[nameKey] = stream
		}
	}
	return res
}

func LookupCrichdStream(streamMap map[string]models.CrichdApiStream, idValue, channelName string) (models.CrichdApiStream, bool) {
	if streamMap == nil {
		return models.CrichdApiStream{}, false
	}
	idKey := strings.ToLower(strings.TrimSpace(idValue))
	nameKey := nonAlphanumericRegex.ReplaceAllString(strings.ToLower(strings.TrimSpace(channelName)), "")

	if s, ok := streamMap[idKey]; ok {
		return s, true
	}
	if s, ok := streamMap[nameKey]; ok {
		return s, true
	}

	var mappedID string
	switch idKey {
	case "willow":
		mappedID = "willowusa"
	case "willowextra":
		mappedID = "willowextra"
	case "starsp":
		mappedID = "star1in"
	case "starsp3":
		mappedID = "starhindi"
	case "crich2":
		mappedID = "skyscric"
	case "skycricket":
		mappedID = "skyscric"
	}
	if mappedID != "" {
		if s, ok := streamMap[mappedID]; ok {
			return s, true
		}
	}

	var mappedName string
	switch {
	case strings.Contains(nameKey, "willowcricket") || strings.Contains(nameKey, "willowhd"):
		mappedName = "willowusa"
	case strings.Contains(nameKey, "willowxtra") || strings.Contains(nameKey, "willowhd2"):
		mappedName = "willowextra"
	case strings.Contains(nameKey, "skycricket") || strings.Contains(nameKey, "skysportscricket"):
		mappedName = "skyscric"
	case strings.Contains(nameKey, "starsports1hindi") || strings.Contains(nameKey, "starsportshindi"):
		mappedName = "starhindi"
	case strings.Contains(nameKey, "starsports1"):
		mappedName = "star1in"
	case strings.Contains(nameKey, "ptvsports"):
		mappedName = "ptvpk"
	case strings.Contains(nameKey, "tensports"):
		mappedName = "tenspk"
	case strings.Contains(nameKey, "asports"):
		mappedName = "asportshd"
	case strings.Contains(nameKey, "tnt1") || strings.Contains(nameKey, "tntsports1"):
		mappedName = "bbtsp1"
	case strings.Contains(nameKey, "tnt2") || strings.Contains(nameKey, "tntsports2"):
		mappedName = "bbtsp2"
	case strings.Contains(nameKey, "tnt3") || strings.Contains(nameKey, "tntsports3"):
		mappedName = "bbtsp3"
	case strings.Contains(nameKey, "tnt4") || strings.Contains(nameKey, "tntsports4") || strings.Contains(nameKey, "tntespn"):
		mappedName = "bbtespn"
	}
	if mappedName != "" {
		if s, ok := streamMap[mappedName]; ok {
			return s, true
		}
	}

	return models.CrichdApiStream{}, false
}

// ParseTapmad parses Tapmad JSON schema
func ParseTapmad(body []byte) []models.TimelineEvent {
	type TapmadItem struct {
		EntityID          string `json:"EntityId"`
		URLSlug           string `json:"UrlSlug"`
		VideoName         string `json:"VideoName"`
		CategoryName      string `json:"CategoryName"`
		StageName         string `json:"StageName"`
		Status            string `json:"Status"`
		EventStartDate    string `json:"EventStartDate"`
		Description       string `json:"Description"`
		ThumbnailTV       string `json:"ThumbnailTV"`
		ThumbnailStandard string `json:"ThumbnailStandard"`
		StreamURL         string `json:"stream_url"`
	}
	type TapmadRoot struct {
		Matches []TapmadItem `json:"Matches"`
	}

	var root TapmadRoot
	if err := json.Unmarshal(body, &root); err != nil {
		return nil
	}

	var list []models.TimelineEvent
	for _, obj := range root.Matches {
		id := obj.EntityID
		if id == "" {
			id = obj.URLSlug
		}
		if id == "" {
			continue
		}

		title := UnescapeHTML(obj.VideoName)
		isLive := strings.EqualFold(obj.Status, "Live")
		startMs := ParseTapmadDate(obj.EventStartDate)
		isDynamic := startMs <= 0
		if startMs <= 0 {
			if isLive {
				startMs = time.Now().UnixMilli()
			} else {
				continue
			}
		}

		poster := obj.ThumbnailTV
		if poster == "" {
			poster = obj.ThumbnailStandard
		}

		var streams []models.EventStream
		if IsStreamPlayable(obj.StreamURL) {
			streams = append(streams, models.EventStream{
				SourceName: "Tapmad",
				StreamURL:  obj.StreamURL,
				Type:       "hls",
				IsM3U:      false,
			})
		}

		cat := obj.CategoryName
		if cat == "" {
			cat = "Sports"
		}

		list = append(list, models.TimelineEvent{
			ID:                 "tapmad_" + id,
			Title:              title,
			Description:        UnescapeHTML(obj.Description),
			StartTimestampMs:   startMs,
			EndTimestampMs:     startMs + DefaultEventDurationMs,
			ChannelID:          "tapmad_" + id,
			Category:           cat,
			League:             obj.StageName,
			Poster:             poster,
			IsLive:             &isLive,
			Streams:            streams,
			IsStartTimeDynamic: isDynamic,
		})
	}
	return list
}

// ParseCrichd parses CricHD and Footy Live schemas
func ParseCrichd(body []byte, sourcePrefix string, crichdStreamMap map[string]models.CrichdApiStream) []models.TimelineEvent {
	type CrichdChannel struct {
		ChannelName string `json:"Channel name"`
		StreamLink  string `json:"Stream link"`
		EmbedLink   string `json:"Embed link"`
	}
	type CrichdItem struct {
		MatchName    string          `json:"match name"`
		Category     string          `json:"Category"`
		TourName     string          `json:"Tour/Group name"`
		Status       string          `json:"Status"`
		StartTime    string          `json:"Start time"`
		Referer      string          `json:"referer"`
		UserAgent    string          `json:"User agent"`
		Team1Name    string          `json:"Team 1 Name"`
		Team1Logo    string          `json:"Team 1 Logo"`
		Team2Name    string          `json:"Team 2 Name"`
		Team2Logo    string          `json:"Team 2 Logo"`
		Channels     []CrichdChannel `json:"Channels"`
	}
	type CrichdRoot struct {
		Matches []CrichdItem `json:"matches"`
	}

	var root CrichdRoot
	if err := json.Unmarshal(body, &root); err != nil {
		return nil
	}

	var list []models.TimelineEvent
	for _, obj := range root.Matches {
		title := UnescapeHTML(obj.MatchName)
		if title == "" {
			continue
		}

		isLive := strings.EqualFold(obj.Status, "LIVE")
		var startMs int64
		if sourcePrefix == "Footy" {
			startMs = ParseIsoDate(obj.StartTime)
		} else {
			startMs = ParseCrichdDate(obj.StartTime)
		}
		isDynamic := strings.EqualFold(obj.StartTime, "N/A") || startMs <= 0
		if startMs <= 0 {
			if isLive {
				startMs = time.Now().UnixMilli()
			} else {
				continue
			}
		}

		var streams []models.EventStream
		for j, ch := range obj.Channels {
			chName := ch.ChannelName
			if chName == "" {
				chName = fmt.Sprintf("Channel %d", j+1)
			}
			streamURL := ch.StreamLink
			if streamURL == "" {
				streamURL = ch.EmbedLink
			}

			if streamURL != "" && strings.HasPrefix(streamURL, "http") {
				resolved := false
				if strings.Contains(streamURL, "playerado.top/embed2.php") || strings.Contains(streamURL, "cricstreams.org/player.php") {
					idParam := "id="
					if idx := strings.Index(streamURL, idParam); idx != -1 {
						idVal := strings.Split(streamURL[idx+len(idParam):], "&")[0]
						if idVal != "" {
							if res, ok := LookupCrichdStream(crichdStreamMap, idVal, chName); ok {
								headers := make(map[string]string)
								if res.Referer != "" {
									headers["Referer"] = res.Referer
								} else if obj.Referer != "" {
									headers["Referer"] = obj.Referer
								}
								if res.Origin != "" {
									headers["Origin"] = res.Origin
								}
								if obj.UserAgent != "" {
									headers["User-Agent"] = obj.UserAgent
								}
								streams = append(streams, models.EventStream{
									SourceName: chName,
									StreamURL:  res.Link,
									Type:       "hls",
									IsM3U:      false,
									Headers:    headers,
								})
								resolved = true
							} else {
								fallbackURL := fmt.Sprintf("https://crichd.johirxofficial.workers.dev/live/%s.m3u8", idVal)
								headers := make(map[string]string)
								if obj.Referer != "" {
									headers["Referer"] = obj.Referer
								}
								if obj.UserAgent != "" {
									headers["User-Agent"] = obj.UserAgent
								}
								streams = append(streams, models.EventStream{
									SourceName: chName,
									StreamURL:  fallbackURL,
									Type:       "hls",
									IsM3U:      false,
									Headers:    headers,
								})
								resolved = true
							}
						}
					}
				}
				if !resolved && IsStreamPlayable(streamURL) {
					headers := make(map[string]string)
					if obj.Referer != "" {
						headers["Referer"] = obj.Referer
					}
					if obj.UserAgent != "" {
						headers["User-Agent"] = obj.UserAgent
					}
					streams = append(streams, models.EventStream{
						SourceName: chName,
						StreamURL:  streamURL,
						Type:       "hls",
						IsM3U:      false,
						Headers:    headers,
					})
				}
			}
		}

		genID := fmt.Sprintf("%s_%d_%d", strings.ToLower(sourcePrefix), HashString(title), startMs)
		if isDynamic {
			genID = fmt.Sprintf("%s_%d_dynamic", strings.ToLower(sourcePrefix), HashString(title))
		}

		cat := obj.Category
		if cat == "" {
			cat = "Sports"
		}

		list = append(list, models.TimelineEvent{
			ID:                 genID,
			Title:              title,
			Description:        "",
			StartTimestampMs:   startMs,
			EndTimestampMs:     startMs + DefaultEventDurationMs,
			ChannelID:          genID,
			Category:           cat,
			League:             obj.TourName,
			HomeTeam:           obj.Team1Name,
			HomeTeamLogo:       obj.Team1Logo,
			AwayTeam:           obj.Team2Name,
			AwayTeamLogo:       obj.Team2Logo,
			IsLive:             &isLive,
			Streams:            streams,
			IsStartTimeDynamic: isDynamic,
		})
	}
	return list
}

// ParseUnifiedMatchesJson parses Bingstream and AXSports JSON formats
func ParseUnifiedMatchesJson(body []byte, sourcePrefix string) []models.TimelineEvent {
	type LinkLiveItem struct {
		DisplayName string `json:"display_name"`
		VideoURL    string `json:"videoURL"`
		StreamLink  string `json:"stream_link"`
		Referer     string `json:"referer"`
	}
	type MatchItem struct {
		Name           string         `json:"name"`
		Category       string         `json:"Category"`
		LeagueName     string         `json:"league_name"`
		LeagueLogo     string         `json:"league_logo"`
		Status         string         `json:"status"`
		IsPlaying      bool           `json:"is_playing"`
		StartAt        int64          `json:"start_at"`
		Timestamp      int64          `json:"timestamp"`
		BdTime         string         `json:"bd_time"`
		Referer        string         `json:"referer"`
		CdnDomain      string         `json:"cdn_domain"`
		LocalteamName  string         `json:"localteam_name"`
		LocalteamLogo  string         `json:"localteam_logo"`
		VisitorteamName string        `json:"visitorteam_name"`
		VisitorteamLogo string        `json:"visitorteam_logo"`
		LinkLive       []LinkLiveItem `json:"link_live"`
	}
	type Root struct {
		Matches []MatchItem `json:"matches"`
	}

	var root Root
	if err := json.Unmarshal(body, &root); err != nil {
		return nil
	}

	var list []models.TimelineEvent
	now := time.Now().UnixMilli()

	for _, obj := range root.Matches {
		title := UnescapeHTML(obj.Name)
		if title == "" {
			continue
		}

		isLive := strings.EqualFold(obj.Status, "LIVE") || obj.IsPlaying
		startAtSec := obj.StartAt
		if startAtSec <= 0 {
			startAtSec = obj.Timestamp
		}

		var startMs int64
		if startAtSec > 0 {
			startMs = startAtSec * 1000
		} else {
			startMs = ParseBingDate(obj.BdTime)
		}

		if isLive && (startMs <= 0 || startMs < now-DefaultEventDurationMs) {
			startMs = now
		}
		isDynamic := startMs <= 0
		if startMs <= 0 {
			continue
		}

		matchRef := obj.Referer
		if matchRef == "" {
			matchRef = obj.CdnDomain
		}

		var streams []models.EventStream
		for j, sObj := range obj.LinkLive {
			serverLabel := sObj.DisplayName
			if serverLabel == "" {
				serverLabel = fmt.Sprintf("Server %d", j+1)
			}
			videoURL := SanitizeStreamURL(sObj.VideoURL)
			streamLink := SanitizeStreamURL(sObj.StreamLink)

			var candidates []string
			if IsStreamPlayable(videoURL) {
				candidates = append(candidates, videoURL)
			}
			if IsStreamPlayable(streamLink) && streamLink != videoURL {
				candidates = append(candidates, streamLink)
			}

			sRef := sObj.Referer
			itemHeaders := make(map[string]string)
			if sRef != "" {
				itemHeaders["Referer"] = sRef
				itemHeaders["Origin"] = strings.TrimSuffix(sRef, "/")
			} else if matchRef != "" {
				itemHeaders["Referer"] = matchRef
				itemHeaders["Origin"] = strings.TrimSuffix(matchRef, "/")
			}

			for _, url := range candidates {
				streams = append(streams, models.EventStream{
					SourceName: serverLabel,
					StreamURL:  url,
					Type:       "hls",
					IsM3U:      false,
					Headers:    itemHeaders,
				})
			}
		}

		genID := fmt.Sprintf("%s_%d_%d", sourcePrefix, HashString(title), startMs)
		if isDynamic {
			genID = fmt.Sprintf("%s_%d_dynamic", sourcePrefix, HashString(title))
		}

		cat := obj.Category
		if cat == "" {
			cat = "Sports"
		}

		list = append(list, models.TimelineEvent{
			ID:                 genID,
			Title:              title,
			Description:        "",
			StartTimestampMs:   startMs,
			EndTimestampMs:     startMs + DefaultEventDurationMs,
			ChannelID:          genID,
			Category:           cat,
			League:             obj.LeagueName,
			HomeTeam:           obj.LocalteamName,
			HomeTeamLogo:       obj.LocalteamLogo,
			AwayTeam:           obj.VisitorteamName,
			AwayTeamLogo:       obj.VisitorteamLogo,
			Poster:             obj.LeagueLogo,
			IsLive:             &isLive,
			Streams:            streams,
			IsStartTimeDynamic: isDynamic,
		})
	}
	return list
}

// ParseSonyLiv parses SonyLIV playlist schema
func ParseSonyLiv(body []byte) []models.TimelineEvent {
	type EmfAttributes struct {
		LandscapeThumb   string `json:"landscape_thumb"`
		Thumbnail        string `json:"thumbnail"`
		VenueName        string `json:"venue_name"`
		BroadcastChannel string `json:"broadcast_channel"`
	}
	type MatchInfo struct {
		ContentID          string        `json:"contentId"`
		Title              string        `json:"title"`
		EpisodeTitle       string        `json:"episodeTitle"`
		Genres             []string      `json:"genres"`
		IsLive             bool          `json:"isLive"`
		IsOnAir            bool          `json:"isOnAir"`
		ContractStartDate  int64         `json:"contractStartDate"`
		OriginalAirDate    int64         `json:"originalAirDate"`
		LongDescription    string        `json:"longDescription"`
		ShortDescription   string        `json:"shortDescription"`
		EmfAttributes      *EmfAttributes `json:"emfAttributes"`
	}
	type SonyItem struct {
		MatchID        string     `json:"match_id"`
		MatchTitle     string     `json:"match_title"`
		SportType      string     `json:"sport_type"`
		HomeTeam       string     `json:"home_team"`
		AwayTeam       string     `json:"away_team"`
		ImageThumbnail string     `json:"image_thumbnail"`
		Status         string     `json:"status"`
		StartTimeBd    string     `json:"start_time_bd"`
		Description    string     `json:"description"`
		Venue          string     `json:"venue"`
		StreamURL      string     `json:"stream_url"`
		MatchInfo      *MatchInfo `json:"match_info"`
	}
	type SonyRoot struct {
		LiveMatches []SonyItem `json:"live_matches"`
	}

	var root SonyRoot
	if err := json.Unmarshal(body, &root); err != nil {
		return nil
	}

	var list []models.TimelineEvent
	now := time.Now().UnixMilli()

	for _, raw := range root.LiveMatches {
		title := ""
		if raw.MatchInfo != nil {
			title = raw.MatchInfo.EpisodeTitle
			if title == "" {
				title = raw.MatchInfo.Title
			}
		}
		if title == "" {
			title = raw.MatchTitle
		}
		title = UnescapeHTML(title)
		if title == "" {
			continue
		}

		matchID := raw.MatchID
		if matchID == "" && raw.MatchInfo != nil {
			matchID = raw.MatchInfo.ContentID
		}
		if matchID == "" {
			matchID = fmt.Sprintf("%d", HashString(title))
		}

		isLive := strings.EqualFold(raw.Status, "LIVE")
		if raw.MatchInfo != nil && (raw.MatchInfo.IsLive || raw.MatchInfo.IsOnAir) {
			isLive = true
		}

		var startMs int64
		if raw.MatchInfo != nil {
			startMs = raw.MatchInfo.ContractStartDate
			if startMs <= 0 {
				startMs = raw.MatchInfo.OriginalAirDate
			}
		}
		if startMs <= 0 {
			startMs = ParseSonyDate(raw.StartTimeBd)
		}

		if isLive && (startMs <= 0 || startMs < now-DefaultEventDurationMs) {
			startMs = now
		}
		isDynamic := startMs <= 0
		if startMs <= 0 {
			continue
		}

		var poster string
		var venue string
		broadcastCh := "SonyLIV"

		if raw.MatchInfo != nil && raw.MatchInfo.EmfAttributes != nil {
			emf := raw.MatchInfo.EmfAttributes
			poster = emf.LandscapeThumb
			if poster == "" {
				poster = emf.Thumbnail
			}
			venue = emf.VenueName
			if emf.BroadcastChannel != "" {
				broadcastCh = emf.BroadcastChannel
			}
		}
		if poster == "" && raw.ImageThumbnail != "N/A" {
			poster = raw.ImageThumbnail
		}
		if venue == "" && raw.Venue != "N/A" {
			venue = raw.Venue
		}

		var streams []models.EventStream
		if IsStreamPlayable(raw.StreamURL) {
			streams = append(streams, models.EventStream{
				SourceName: broadcastCh,
				StreamURL:  raw.StreamURL,
				Type:       "hls",
				IsM3U:      false,
			})
		}

		genID := fmt.Sprintf("sony_%s", matchID)
		if isDynamic {
			genID = fmt.Sprintf("sony_%s_dynamic", matchID)
		}

		cat := "Sports"
		if raw.MatchInfo != nil && len(raw.MatchInfo.Genres) > 0 {
			cat = raw.MatchInfo.Genres[0]
		} else if raw.SportType != "" {
			cat = raw.SportType
		}

		list = append(list, models.TimelineEvent{
			ID:                 genID,
			Title:              title,
			Description:        UnescapeHTML(raw.Description),
			StartTimestampMs:   startMs,
			EndTimestampMs:     startMs + DefaultEventDurationMs,
			ChannelID:          genID,
			Category:           cat,
			HomeTeam:           raw.HomeTeam,
			AwayTeam:           raw.AwayTeam,
			Poster:             poster,
			Venue:              venue,
			IsLive:             &isLive,
			Streams:            streams,
			IsStartTimeDynamic: isDynamic,
		})
	}
	return list
}

// ParseWillow parses Willow and Prime Video sports JSON schemas
func ParseWillow(body []byte) []models.TimelineEvent {
	type WillowItem struct {
		MatchID       string                 `json:"match_id"`
		Title         string                 `json:"title"`
		CoverImage    string                 `json:"cover_image"`
		Status        string                 `json:"status"`
		TimeStr       string                 `json:"time"`
		Synopsis      string                 `json:"synopsis"`
		DrmKey        string                 `json:"drm_key"`
		StreamURLObj  map[string]string      `json:"stream_url"`
		StreamAlpha   map[string]string      `json:"stream_url_alpha"`
		StreamBravo   map[string]string      `json:"stream_url_bravo"`
	}
	type WillowRoot struct {
		Matches []WillowItem `json:"Matches"`
	}

	var root WillowRoot
	if err := json.Unmarshal(body, &root); err != nil {
		return nil
	}

	var list []models.TimelineEvent
	now := time.Now().UnixMilli()

	for _, obj := range root.Matches {
		if obj.MatchID == "" {
			continue
		}
		title := UnescapeHTML(obj.Title)
		isLive := strings.EqualFold(obj.Status, "LIVE")

		startMs := ParseWillowDate(obj.TimeStr)
		if isLive && (startMs <= 0 || startMs < now-DefaultEventDurationMs) {
			startMs = now
		}
		isDynamic := startMs <= 0
		if startMs <= 0 {
			continue
		}

		var drmConfig *models.DrmConfig
		if obj.DrmKey != "" {
			drmConfig = &models.DrmConfig{ClearKey: obj.DrmKey}
		}

		var streams []models.EventStream
		for serverName, url := range obj.StreamURLObj {
			san := SanitizeStreamURL(url)
			if IsStreamPlayable(san) {
				sType := "hls"
				if strings.Contains(strings.ToLower(san), ".mpd") {
					sType = "dash"
				}
				streams = append(streams, models.EventStream{
					SourceName: serverName,
					StreamURL:  san,
					Type:       sType,
					IsM3U:      false,
					Headers:    GetStreamHeadersForURL(san, nil),
					DRM:        drmConfig,
				})
			}
		}
		for serverName, url := range obj.StreamAlpha {
			san := SanitizeStreamURL(url)
			if IsStreamPlayable(san) {
				sType := "hls"
				if strings.Contains(strings.ToLower(san), ".mpd") {
					sType = "dash"
				}
				streams = append(streams, models.EventStream{
					SourceName: serverName,
					StreamURL:  san,
					Type:       sType,
					IsM3U:      false,
					DRM:        drmConfig,
				})
			}
		}
		for serverName, url := range obj.StreamBravo {
			san := SanitizeStreamURL(url)
			if IsStreamPlayable(san) {
				sType := "hls"
				if strings.Contains(strings.ToLower(san), ".mpd") {
					sType = "dash"
				}
				streams = append(streams, models.EventStream{
					SourceName: fmt.Sprintf("Bravo - %s", serverName),
					StreamURL:  san,
					Type:       sType,
					IsM3U:      false,
					DRM:        drmConfig,
				})
			}
		}

		genID := fmt.Sprintf("willow_%s", obj.MatchID)
		if isDynamic {
			genID = fmt.Sprintf("willow_%s_dynamic", obj.MatchID)
		}

		list = append(list, models.TimelineEvent{
			ID:                 genID,
			Title:              title,
			Description:        UnescapeHTML(obj.Synopsis),
			StartTimestampMs:   startMs,
			EndTimestampMs:     startMs + DefaultEventDurationMs,
			ChannelID:          genID,
			Category:           "Sports",
			Poster:             obj.CoverImage,
			IsLive:             &isLive,
			Streams:            streams,
			IsStartTimeDynamic: isDynamic,
		})
	}
	return list
}

// ParseSportsData parses Sports Data schema
func ParseSportsData(body []byte) []models.TimelineEvent {
	type StreamItem struct {
		StreamURL   string `json:"stream_url"`
		Name        string `json:"name"`
		ChannelName string `json:"channel_name"`
		DrmKey      string `json:"drm_key"`
		Kid         string `json:"kid"`
		Key         string `json:"key"`
		Referer     string `json:"referer"`
		Origin      string `json:"origin"`
		UserAgent   string `json:"user_agent"`
		Cookie      string `json:"cookie"`
	}
	type EventInfo struct {
		TeamA     string `json:"teamA"`
		TeamB     string `json:"teamB"`
		TeamAFlag string `json:"teamAFlag"`
		TeamBFlag string `json:"teamBFlag"`
		EventName string `json:"eventName"`
		StartTime string `json:"startTime"`
	}
	type SportsDataItem struct {
		Status    string       `json:"status"`
		EventName string       `json:"event_name"`
		Category  string       `json:"Category"`
		EventInfo *EventInfo   `json:"eventInfo"`
		Streams   []StreamItem `json:"streams"`
	}
	type SportsDataRoot struct {
		Matches []SportsDataItem `json:"matches"`
	}

	var root SportsDataRoot
	if err := json.Unmarshal(body, &root); err != nil {
		return nil
	}

	var list []models.TimelineEvent
	now := time.Now().UnixMilli()

	for _, obj := range root.Matches {
		if strings.EqualFold(obj.Status, "FINISHED") {
			continue
		}
		isLive := strings.EqualFold(obj.Status, "LIVE")

		eventNameRaw := strings.TrimSpace(obj.EventName)
		var homeTeam, awayTeam, homeLogo, awayLogo, leagueName, startTimeStr string
		if obj.EventInfo != nil {
			homeTeam = strings.TrimSpace(obj.EventInfo.TeamA)
			awayTeam = strings.TrimSpace(obj.EventInfo.TeamB)
			homeLogo = strings.TrimSpace(obj.EventInfo.TeamAFlag)
			awayLogo = strings.TrimSpace(obj.EventInfo.TeamBFlag)
			leagueName = strings.TrimSpace(obj.EventInfo.EventName)
			startTimeStr = obj.EventInfo.StartTime
		}

		title := ""
		if eventNameRaw != "" {
			title = UnescapeHTML(eventNameRaw)
		} else if homeTeam != "" && awayTeam != "" {
			title = fmt.Sprintf("%s vs %s", homeTeam, awayTeam)
		} else if leagueName != "" {
			title = leagueName
		}
		if title == "" {
			continue
		}

		startMs := ParseSportsDataDate(startTimeStr)
		if isLive && (startMs <= 0 || startMs < now-DefaultEventDurationMs) {
			startMs = now
		}
		isDynamic := startMs <= 0
		if startMs <= 0 {
			continue
		}

		var streams []models.EventStream
		for j, sObj := range obj.Streams {
			rawURL := sObj.StreamURL
			streamName := strings.TrimSpace(sObj.ChannelName)
			if streamName == "" {
				streamName = strings.TrimSpace(sObj.Name)
			}
			if streamName == "" {
				streamName = fmt.Sprintf("Server %d", j+1)
			}
			if rawURL == "" || !strings.HasPrefix(rawURL, "http") {
				continue
			}

			cleanURL := rawURL
			headersMap := make(map[string]string)

			if strings.Contains(rawURL, "|") {
				parts := strings.SplitN(rawURL, "|", 2)
				cleanURL = parts[0]
				if len(parts) > 1 {
					headerPairs := strings.Split(parts[1], "&")
					for _, pair := range headerPairs {
						kv := strings.SplitN(pair, "=", 2)
						if len(kv) == 2 && strings.TrimSpace(kv[0]) != "" && strings.TrimSpace(kv[1]) != "" {
							headersMap[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
						}
					}
				}
			}

			if sObj.Referer != "" {
				headersMap["Referer"] = sObj.Referer
			}
			if sObj.Origin != "" {
				headersMap["Origin"] = sObj.Origin
			}
			if sObj.UserAgent != "" {
				headersMap["User-Agent"] = sObj.UserAgent
			}
			if sObj.Cookie != "" {
				headersMap["Cookie"] = sObj.Cookie
			}

			cleanURL = SanitizeStreamURL(cleanURL)
			if IsStreamPlayable(cleanURL) {
				var drmConfig *models.DrmConfig
				if sObj.DrmKey != "" {
					drmConfig = &models.DrmConfig{ClearKey: sObj.DrmKey}
				} else if sObj.Kid != "" && sObj.Key != "" {
					drmConfig = &models.DrmConfig{ClearKey: fmt.Sprintf("%s:%s", strings.TrimSpace(sObj.Kid), strings.TrimSpace(sObj.Key))}
				}
				sType := "hls"
				if strings.Contains(strings.ToLower(cleanURL), ".mpd") {
					sType = "dash"
				}
				finalHeaders := GetStreamHeadersForURL(cleanURL, headersMap)
				streams = append(streams, models.EventStream{
					SourceName: streamName,
					StreamURL:  cleanURL,
					Type:       sType,
					IsM3U:      false,
					Headers:    finalHeaders,
					DRM:        drmConfig,
				})
			}
		}

		genID := fmt.Sprintf("sportsdata_%d_%d", HashString(title), startMs)
		cat := obj.Category
		if cat == "" {
			cat = "Sports"
		}

		list = append(list, models.TimelineEvent{
			ID:                 genID,
			Title:              title,
			Description:        "",
			StartTimestampMs:   startMs,
			EndTimestampMs:     startMs + DefaultEventDurationMs,
			ChannelID:          genID,
			Category:           cat,
			League:             leagueName,
			HomeTeam:           homeTeam,
			HomeTeamLogo:       homeLogo,
			AwayTeam:           awayTeam,
			AwayTeamLogo:       awayLogo,
			IsLive:             &isLive,
			Streams:            streams,
			IsStartTimeDynamic: isDynamic,
		})
	}
	return list
}
