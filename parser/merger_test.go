package parser

import (
	"testing"
	"github.com/neelniloy/lstv-event-parser/models"
)

func TestGetMatchKey(t *testing.T) {
	tests := []struct {
		name     string
		event    models.TimelineEvent
		expected string
	}{
		{
			name: "Exact Team Names",
			event: models.TimelineEvent{
				HomeTeam: "Manchester United",
				AwayTeam: "Arsenal",
			},
			expected: "arsenal_vs_manchesterunited",
		},
		{
			name: "Team Alias Normalization",
			event: models.TimelineEvent{
				HomeTeam: "Man Utd",
				AwayTeam: "Arsenal",
			},
			expected: "arsenal_vs_manchesterunited",
		},
		{
			name: "Country Alias Normalization",
			event: models.TimelineEvent{
				HomeTeam: "IND",
				AwayTeam: "PAK",
			},
			expected: "india_vs_pakistan",
		},
		{
			name: "Title Fallback Parsing",
			event: models.TimelineEvent{
				Title: "Real Madrid vs Barca",
			},
			expected: "barcelona_vs_realmadrid",
		},
		{
			name: "Country Tag in Parentheses Normalization",
			event: models.TimelineEvent{
				HomeTeam: "Al Ain FC (UAE)",
				AwayTeam: "Tractor FC (IRN)",
			},
			expected: "alain_vs_tractor",
		},
		{
			name: "French Club Aliases Normalization",
			event: models.TimelineEvent{
				HomeTeam: "PSG",
				AwayTeam: "OL",
			},
			expected: "olympiquelyonnais_vs_psg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := GetMatchKey(&tt.event)
			if key != tt.expected {
				t.Errorf("GetMatchKey() = %v, want %v", key, tt.expected)
			}
		})
	}
}

func TestIsPlaceholderEvent(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		homeTeam string
		awayTeam string
		expected bool
	}{
		{
			name:     "Identical Home and Away Teams",
			title:    "MotorGP Vs MotorGP",
			homeTeam: "MotorGP",
			awayTeam: "MotorGP",
			expected: true,
		},
		{
			name:     "Generic Evento Placeholder",
			title:    "TENNIS | EVENTO  Vs TENNIS | EVENTO",
			homeTeam: "TENNIS | EVENTO",
			awayTeam: "TENNIS | EVENTO",
			expected: true,
		},
		{
			name:     "Generic Sports Live Placeholder",
			title:    "SPORTS LIVE Vs SPORTS LIVE",
			homeTeam: "SPORTS LIVE",
			awayTeam: "SPORTS LIVE",
			expected: true,
		},
		{
			name:     "Tennis Court Name Placeholder",
			title:    "Centre Court",
			homeTeam: "",
			awayTeam: "",
			expected: true,
		},
		{
			name:     "Real Match Event",
			title:    "Arsenal vs Chelsea",
			homeTeam: "Arsenal",
			awayTeam: "Chelsea",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsPlaceholderEvent(tt.title, tt.homeTeam, tt.awayTeam)
			if got != tt.expected {
				t.Errorf("IsPlaceholderEvent(%q, %q, %q) = %v, want %v", tt.title, tt.homeTeam, tt.awayTeam, got, tt.expected)
			}
		})
	}
}

func TestMergeAndDeduplicate(t *testing.T) {
	events := []models.TimelineEvent{
		{
			ID:               "1",
			Title:            "Man Utd vs Arsenal",
			HomeTeam:         "Man Utd",
			AwayTeam:         "Arsenal",
			StartTimestampMs: 1000000,
			EndTimestampMs:   1000000 + 3600000,
			Streams: []models.EventStream{
				{SourceName: "Stream 1", StreamURL: "http://example.com/s1.m3u8"},
			},
		},
		{
			ID:               "2",
			Title:            "Manchester United vs Arsenal",
			HomeTeam:         "Manchester United",
			AwayTeam:         "Arsenal",
			StartTimestampMs: 1005000,
			EndTimestampMs:   1005000 + 3600000,
			Streams: []models.EventStream{
				{SourceName: "Stream 2", StreamURL: "http://example.com/s2.m3u8"},
			},
		},
	}

	merged := MergeAndDeduplicate(events, 1000000)
	if len(merged) != 1 {
		t.Fatalf("Expected 1 merged event, got %d", len(merged))
	}
	if len(merged[0].Streams) != 2 {
		t.Errorf("Expected 2 combined streams, got %d", len(merged[0].Streams))
	}
}

func TestIsMatchingEventTournamentDayDoesNotSwallowMatches(t *testing.T) {
	day := models.TimelineEvent{Title: "Hong Kong Sixes - Day 1", StartTimestampMs: 1_000_000}
	match := models.TimelineEvent{Title: "Hong Kong Sixes 2026 - 6th Match - Bangladesh vs Sri Lanka", StartTimestampMs: 1_000_000}
	if IsMatchingEvent(&day, &match) {
		t.Errorf("tournament day event should not merge with an individual match")
	}

	fancode := models.TimelineEvent{Title: "Bangladesh vs Sri Lanka", HomeTeam: "Bangladesh", AwayTeam: "Sri Lanka", StartTimestampMs: 1_000_000}
	if !IsMatchingEvent(&fancode, &match) {
		t.Errorf("same match from two sources should merge")
	}
}

func TestRemoveSharedStreams(t *testing.T) {
	shared := models.EventStream{StreamURL: "https://cdn/tournament.mpd"}
	events := []models.TimelineEvent{
		{Title: "Bangladesh vs Sri Lanka", Streams: []models.EventStream{{StreamURL: "https://cdn/bansl.mpd"}, shared}},
		{Title: "India vs Pakistan", Streams: []models.EventStream{shared}},
	}
	RemoveSharedStreams(events)

	if len(events[0].Streams) != 1 || events[0].Streams[0].StreamURL != "https://cdn/bansl.mpd" {
		t.Errorf("shared stream should be removed when a dedicated one exists, got %+v", events[0].Streams)
	}
	if len(events[1].Streams) != 1 {
		t.Errorf("shared stream should be kept when it is the only one, got %+v", events[1].Streams)
	}
}
