package parser

import (
	"testing"
	"github.com/neelniloy/lstv-event-parser/models"
)

func TestGetSportCategory(t *testing.T) {
	tests := []struct {
		name     string
		event    models.TimelineEvent
		expected string
	}{
		{
			name: "Cricket IPL",
			event: models.TimelineEvent{
				Title: "RCB vs CSK IPL 2024",
			},
			expected: "Cricket",
		},
		{
			name: "Football Premier League",
			event: models.TimelineEvent{
				Title: "Chelsea vs Liverpool Premier League",
			},
			expected: "Football",
		},
		{
			name: "Golf PGA Tour",
			event: models.TimelineEvent{
				Title: "PGA Tour Championship Day 1",
			},
			expected: "Golf",
		},
		{
			name: "Baseball MLB",
			event: models.TimelineEvent{
				Title: "Yankees vs Red Sox MLB",
			},
			expected: "Baseball",
		},
		{
			name: "Esports Valorant",
			event: models.TimelineEvent{
				Title: "Valorant Champions Tour Grand Final",
			},
			expected: "Esports",
		},
		{
			name: "NFL Super Bowl",
			event: models.TimelineEvent{
				Title: "Super Bowl LVIII",
			},
			expected: "American Football",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cat := GetSportCategory(&tt.event)
			if cat != tt.expected {
				t.Errorf("GetSportCategory() = %v, want %v", cat, tt.expected)
			}
		})
	}
}

func TestInferLeague(t *testing.T) {
	ev := models.TimelineEvent{
		Title: "Arsenal vs Chelsea Premier League Match",
	}
	GetSportCategory(&ev)
	if ev.League != "Premier League" {
		t.Errorf("Expected inferred league Premier League, got %s", ev.League)
	}
}
