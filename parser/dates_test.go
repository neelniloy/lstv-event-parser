package parser

import (
	"testing"
)

func TestParseWillowDate(t *testing.T) {
	ts := ParseWillowDate("Tomorrow 3:00 PM BDT")
	if ts <= 0 {
		t.Errorf("Expected positive timestamp for Willow relative date, got %d", ts)
	}

	ts2 := ParseWillowDate("Mon, Aug 19 8 PM")
	if ts2 <= 0 {
		t.Errorf("Expected positive timestamp for Willow fixed date, got %d", ts2)
	}

	ts3 := ParseWillowDate("Live at 5:30 PM BDT")
	if ts3 <= 0 {
		t.Errorf("Expected positive timestamp for Willow 'Live at' date, got %d", ts3)
	}

	ts4 := ParseWillowDate("LIVE NOW")
	if ts4 <= 0 {
		t.Errorf("Expected positive timestamp for Willow 'LIVE NOW', got %d", ts4)
	}
}

func TestParseSportsDataDate(t *testing.T) {
	ts := ParseSportsDataDate("04/10/2026 09:35:00 AM")
	if ts <= 0 {
		t.Errorf("Expected positive timestamp for SportsData date with seconds, got %d", ts)
	}

	ts2 := ParseSportsDataDate("04-10-2026 15:04:05")
	if ts2 <= 0 {
		t.Errorf("Expected positive timestamp for SportsData 24h format, got %d", ts2)
	}
}

func TestParseIsoDate(t *testing.T) {
	ts := ParseIsoDate("2026-08-16T18:00:00Z")
	if ts <= 0 {
		t.Errorf("Expected positive timestamp for ISO date, got %d", ts)
	}
}

