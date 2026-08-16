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
}

func TestParseIsoDate(t *testing.T) {
	ts := ParseIsoDate("2026-08-16T18:00:00Z")
	if ts <= 0 {
		t.Errorf("Expected positive timestamp for ISO date, got %d", ts)
	}
}
