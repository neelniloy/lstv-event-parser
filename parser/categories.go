package parser

import (
	"strings"
	"github.com/neelniloy/lstv-event-parser/models"
)

func GetSportCategory(event *models.TimelineEvent) string {
	titleLower := strings.ToLower(event.Title)
	catLower := strings.TrimSpace(strings.ToLower(event.Category))
	leagueLower := strings.ToLower(event.League)

	// 1. Motorsport
	motorsportKeywords := []string{"formula 1", "formula1", "f1", "motogp", "nascar", "indycar", "rally", "motorsport"}
	for _, kw := range motorsportKeywords {
		if strings.Contains(catLower, kw) || strings.Contains(titleLower, kw) || strings.Contains(leagueLower, kw) {
			return "Motorsport"
		}
	}

	// 2. Combat Sports
	combatKeywords := []string{"ufc", "mma", "wwe", "wrestling", "boxing", "one championship", "bellator"}
	for _, kw := range combatKeywords {
		if strings.Contains(catLower, kw) || strings.Contains(titleLower, kw) || strings.Contains(leagueLower, kw) {
			return "Combat Sports"
		}
	}

	// 3. Tennis
	tennisKeywords := []string{"tennis", "wimbledon", "us open", "french open", "australian open", "atp", "wta", "roland garros"}
	for _, kw := range tennisKeywords {
		if strings.Contains(catLower, kw) || strings.Contains(titleLower, kw) || strings.Contains(leagueLower, kw) {
			return "Tennis"
		}
	}

	// 4. Basketball
	basketballKeywords := []string{"basketball", "nba", "euroleague", "fiba"}
	for _, kw := range basketballKeywords {
		if strings.Contains(catLower, kw) || strings.Contains(titleLower, kw) || strings.Contains(leagueLower, kw) {
			return "Basketball"
		}
	}

	// 5. Explicit Category Match
	if strings.Contains(catLower, "cricket") {
		return "Cricket"
	}
	if strings.Contains(catLower, "football") || strings.Contains(catLower, "soccer") {
		return "Football"
	}

	// 6. Cricket Keywords & Leagues
	cricketKeywords := []string{
		"cricket", "t20", "odi", "test match", "test series", "t10", "ipl", "bpl", "psl", "cpl",
		"big bash", "blast", "ashes", "tri-series", "test", "tour of", "the hundred", "hundred",
		"super giants", "phoenix", "invincibles", "superchargers", "welsh fire", "southern brave",
		"trent rockets", "london spirit", "manchester originals", "mlc", "major league cricket",
		"vitality blast", "county championship",
	}
	for _, kw := range cricketKeywords {
		if strings.Contains(titleLower, kw) || strings.Contains(catLower, kw) || strings.Contains(leagueLower, kw) {
			return "Cricket"
		}
	}

	// 7. Football Keywords
	footballKeywords := []string{
		"football", "soccer", "laliga", "serie a", "bundesliga", "ligue 1",
		"uefa", "champions league", "europa league", "copa america", "euro 202", "fifa", "la liga",
		"mls", "real madrid", "barcelona", "bayern", "juventus", "psg", "premier league", "chelsea",
		"manchester united", "manchester city", "man utd", "man city", "liverpool", "arsenal", "tottenham",
		"ac milan", "inter milan", "atletico",
	}
	for _, kw := range footballKeywords {
		if strings.Contains(titleLower, kw) || strings.Contains(catLower, kw) || strings.Contains(leagueLower, kw) {
			return "Football"
		}
	}

	// 8. World Cup / Tournament Context
	if strings.Contains(titleLower, "world cup") || strings.Contains(catLower, "world cup") || strings.Contains(leagueLower, "world cup") {
		if strings.Contains(titleLower, "fifa") || strings.Contains(titleLower, "soccer") || strings.Contains(titleLower, "football") {
			return "Football"
		}
		cricketTeams := []string{"india", "pakistan", "bangladesh", "england", "australia", "new zealand", "south africa", "west indies", "sri lanka", "afghanistan"}
		for _, team := range cricketTeams {
			if strings.Contains(titleLower, team) {
				return "Cricket"
			}
		}
		return "Football"
	}

	if strings.Contains(titleLower, "asia cup") || strings.Contains(catLower, "asia cup") || strings.Contains(leagueLower, "asia cup") {
		return "Cricket"
	}

	if event.Category != "" && !strings.EqualFold(event.Category, "Sports") {
		return strings.TrimSpace(event.Category)
	}

	return "Other Sports"
}
