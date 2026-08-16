package models

// DrmConfig defines ClearKey DRM configuration if present
type DrmConfig struct {
	ClearKey string `json:"clearkey,omitempty"`
}

// EventStream defines playable stream metadata
type EventStream struct {
	SourceName string            `json:"sourceName"`
	StreamURL  string            `json:"streamUrl"`
	Type       string            `json:"type"`
	IsM3U      bool              `json:"isM3u"`
	Headers    map[string]string `json:"headers,omitempty"`
	DRM        *DrmConfig        `json:"drm,omitempty"`
}

// TimelineEvent is the unified live sports event output model
type TimelineEvent struct {
	ID                  string        `json:"id"`
	Title               string        `json:"title"`
	Description         string        `json:"description"`
	StartTimestampMs    int64         `json:"startTimestampMs"`
	EndTimestampMs      int64         `json:"endTimestampMs"`
	ChannelID           string        `json:"channelId"`
	Category            string        `json:"category"`
	League              string        `json:"league,omitempty"`
	LeagueLogo          string        `json:"leagueLogo,omitempty"`
	HomeTeam            string        `json:"homeTeam,omitempty"`
	HomeTeamLogo        string        `json:"homeTeamLogo,omitempty"`
	AwayTeam            string        `json:"awayTeam,omitempty"`
	AwayTeamLogo        string        `json:"awayTeamLogo,omitempty"`
	IsLive              *bool         `json:"isLive,omitempty"`
	Venue               string        `json:"venue,omitempty"`
	Poster              string        `json:"poster,omitempty"`
	Streams             []EventStream `json:"streams,omitempty"`
	IsStartTimeDynamic  bool          `json:"isStartTimeDynamic"`
}

// CrichdApiItem represents entries from the CricHD channel stream lookup API
type CrichdApiItem struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Link    string `json:"link"`
	Referer string `json:"referer"`
	Origin  string `json:"origin"`
}

// CrichdApiStream holds resolved stream details
type CrichdApiStream struct {
	Link    string
	Referer string
	Origin  string
}
