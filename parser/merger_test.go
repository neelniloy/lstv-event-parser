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
