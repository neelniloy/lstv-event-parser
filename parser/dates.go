package parser

import (
	"strings"
	"time"
)

var (
	gmtPlus6 = time.FixedZone("GMT+6", 6*3600)
)

func ParseTapmadDate(dateStr string) int64 {
	if strings.TrimSpace(dateStr) == "" {
		return 0
	}
	t, err := time.ParseInLocation("2006-01-02 15:04:05", strings.TrimSpace(dateStr), gmtPlus6)
	if err != nil {
		return 0
	}
	return t.UnixMilli()
}

func ParseCrichdDate(dateStr string) int64 {
	if strings.TrimSpace(dateStr) == "" {
		return 0
	}
	cleaned := strings.ReplaceAll(strings.TrimSpace(dateStr), " at ", " ")
	formats := []string{
		"January 2, 2006 3:04 PM MST",
		"January 2, 2006 3 PM MST",
		"January 2, 2006 15:04 MST",
		"January 2, 2006 3:04 PM",
		"January 2, 2006 3 PM",
	}
	for _, fmt := range formats {
		t, err := time.ParseInLocation(fmt, cleaned, time.UTC)
		if err == nil {
			return t.UnixMilli()
		}
	}
	return 0
}

func ParseIsoDate(dateStr string) int64 {
	if strings.TrimSpace(dateStr) == "" {
		return 0
	}
	formats := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05Z",
	}
	for _, fmt := range formats {
		t, err := time.Parse(fmt, strings.TrimSpace(dateStr))
		if err == nil {
			return t.UnixMilli()
		}
	}
	return 0
}

func ParseBingDate(dateStr string) int64 {
	if strings.TrimSpace(dateStr) == "" {
		return 0
	}
	if strings.EqualFold(strings.TrimSpace(dateStr), "Live Now") {
		return time.Now().UnixMilli()
	}
	t, err := time.ParseInLocation("3:04 PM 02-01-2006", strings.TrimSpace(dateStr), gmtPlus6)
	if err != nil {
		return 0
	}
	return t.UnixMilli()
}

func ParseSonyDate(dateStr string) int64 {
	if strings.TrimSpace(dateStr) == "" {
		return 0
	}
	formats := []string{
		"03:04:05 PM 02-01-2006",
		"3:04:05 PM 02-01-2006",
	}
	for _, fmt := range formats {
		t, err := time.ParseInLocation(fmt, strings.TrimSpace(dateStr), gmtPlus6)
		if err == nil {
			return t.UnixMilli()
		}
	}
	return 0
}

func ParseSportsDataDate(dateStr string) int64 {
	if strings.TrimSpace(dateStr) == "" {
		return 0
	}
	formats := []string{
		"02-01-2006 03:04 PM",
		"02-01-2006 3:04 PM",
		"02-01-2006 15:04",
		"02/01/2006 03:04 PM",
		"02/01/2006 3:04 PM",
		"02/01/2006 15:04",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		time.RFC3339Nano,
		time.RFC3339,
	}
	for _, fmt := range formats {
		t, err := time.ParseInLocation(fmt, strings.TrimSpace(dateStr), gmtPlus6)
		if err == nil {
			return t.UnixMilli()
		}
	}
	return 0
}

func ParseWillowDate(dateStr string) int64 {
	if strings.TrimSpace(dateStr) == "" {
		return 0
	}
	cleaned := strings.TrimSpace(dateStr)
	now := time.Now().In(gmtPlus6)

	if strings.HasPrefix(strings.ToLower(cleaned), "tomorrow") {
		timePart := strings.TrimSpace(strings.ReplaceAll(cleaned[8:], "BDT", ""))
		layout := "3:04 PM"
		if !strings.Contains(timePart, ":") {
			layout = "3 PM"
		}
		t, err := time.ParseInLocation(layout, timePart, gmtPlus6)
		if err == nil {
			tomorrow := now.AddDate(0, 0, 1)
			result := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), t.Hour(), t.Minute(), 0, 0, gmtPlus6)
			return result.UnixMilli()
		}
	}

	if !strings.Contains(cleaned, ",") && (strings.Contains(cleaned, "AM BDT") || strings.Contains(cleaned, "PM BDT")) {
		timePart := strings.TrimSpace(strings.ReplaceAll(cleaned, "BDT", ""))
		layout := "3:04 PM"
		if !strings.Contains(timePart, ":") {
			layout = "3 PM"
		}
		t, err := time.ParseInLocation(layout, timePart, gmtPlus6)
		if err == nil {
			result := time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, gmtPlus6)
			return result.UnixMilli()
		}
	}

	if strings.Contains(cleaned, ",") {
		timeStr := cleaned
		if strings.HasSuffix(cleaned, "BDT") {
			timeStr = strings.TrimSpace(strings.ReplaceAll(cleaned, "BDT", ""))
		}
		formats := []string{
			"Mon, Jan 2 3:04 PM",
			"Mon, Jan 2 3 PM",
		}
		for _, fmt := range formats {
			t, err := time.ParseInLocation(fmt, timeStr, gmtPlus6)
			if err == nil {
				result := time.Date(now.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, gmtPlus6)
				return result.UnixMilli()
			}
		}
	}

	return 0
}
