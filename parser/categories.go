package parser

import (
	"strings"
	"github.com/neelniloy/lstv-event-parser/models"
)

func GetSportCategory(event *models.TimelineEvent) string {
	titleLower := strings.ToLower(event.Title)
	catLower := strings.TrimSpace(strings.ToLower(event.Category))
	leagueLower := strings.ToLower(event.League)

	// Infer league if missing
	if event.League == "" {
		event.League = InferLeague(titleLower, catLower)
		leagueLower = strings.ToLower(event.League)
	}

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
	tennisKeywords := []string{"tennis", "wimbledon", "us open tennis", "french open", "australian open", "atp", "wta", "roland garros"}
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

	// 5. Golf
	golfKeywords := []string{"golf", "pga tour", "dp world tour", "liv golf", "masters tournament"}
	for _, kw := range golfKeywords {
		if strings.Contains(catLower, kw) || strings.Contains(titleLower, kw) || strings.Contains(leagueLower, kw) {
			return "Golf"
		}
	}

	// 6. Baseball
	baseballKeywords := []string{"baseball", "mlb"}
	for _, kw := range baseballKeywords {
		if strings.Contains(catLower, kw) || strings.Contains(titleLower, kw) || strings.Contains(leagueLower, kw) {
			return "Baseball"
		}
	}

	// 7. Rugby
	rugbyKeywords := []string{"rugby", "six nations", "nrl", "super rugby"}
	for _, kw := range rugbyKeywords {
		if strings.Contains(catLower, kw) || strings.Contains(titleLower, kw) || strings.Contains(leagueLower, kw) {
			return "Rugby"
		}
	}

	// 8. Badminton & Volleyball & Ice Hockey & Esports & American Football
	if strings.Contains(catLower, "badminton") || strings.Contains(titleLower, "badminton") || strings.Contains(titleLower, "bwf") {
		return "Badminton"
	}
	if strings.Contains(catLower, "volleyball") || strings.Contains(titleLower, "volleyball") || strings.Contains(titleLower, "vnl") {
		return "Volleyball"
	}
	if strings.Contains(catLower, "ice hockey") || strings.Contains(titleLower, "ice hockey") || strings.Contains(titleLower, "nhl") {
		return "Ice Hockey"
	}
	if strings.Contains(catLower, "esports") || strings.Contains(titleLower, "valorant") || strings.Contains(titleLower, "cs:go") || strings.Contains(titleLower, "dota 2") {
		return "Esports"
	}
	if strings.Contains(catLower, "nfl") || strings.Contains(titleLower, "nfl") || strings.Contains(titleLower, "super bowl") {
		return "American Football"
	}

	// 9. Explicit Category Match
	if strings.Contains(catLower, "cricket") {
		return "Cricket"
	}
	if strings.Contains(catLower, "football") || strings.Contains(catLower, "soccer") {
		return "Football"
	}

	// 10. Cricket Keywords & Leagues
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

	// 11. Football Keywords
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

	// 12. World Cup / Tournament Context
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

func InferLeague(titleLower, catLower string) string {
	switch {
	case strings.Contains(titleLower, "premier league") || strings.Contains(titleLower, "epl"):
		return "Premier League"
	case strings.Contains(titleLower, "la liga") || strings.Contains(titleLower, "laliga"):
		return "La Liga"
	case strings.Contains(titleLower, "serie a"):
		return "Serie A"
	case strings.Contains(titleLower, "bundesliga"):
		return "Bundesliga"
	case strings.Contains(titleLower, "ligue 1"):
		return "Ligue 1"
	case strings.Contains(titleLower, "champions league") || strings.Contains(titleLower, "ucl"):
		return "UEFA Champions League"
	case strings.Contains(titleLower, "europa league") || strings.Contains(titleLower, "uel"):
		return "UEFA Europa League"
	case strings.Contains(titleLower, "ipl"):
		return "Indian Premier League"
	case strings.Contains(titleLower, "bpl"):
		return "Bangladesh Premier League"
	case strings.Contains(titleLower, "psl"):
		return "Pakistan Super League"
	case strings.Contains(titleLower, "cpl"):
		return "Caribbean Premier League"
	case strings.Contains(titleLower, "nba"):
		return "NBA"
	case strings.Contains(titleLower, "nfl"):
		return "NFL"
	case strings.Contains(titleLower, "mlb"):
		return "MLB"
	case strings.Contains(titleLower, "formula 1") || strings.Contains(titleLower, "f1"):
		return "Formula 1"
	case strings.Contains(titleLower, "ufc"):
		return "UFC"
	}
	return ""
}

func GetSportMaxDurationMs(category, title string) int64 {
	catLower := strings.ToLower(category)
	titleLower := strings.ToLower(title)

	switch {
	case catLower == "cricket":
		if strings.Contains(titleLower, "test") || strings.Contains(titleLower, "day ") {
			return 5 * 24 * 60 * 60 * 1000 // 5 days for Test Cricket
		}
		if strings.Contains(titleLower, "odi") || strings.Contains(titleLower, "one day") {
			return 12 * 60 * 60 * 1000 // 12 hours for ODI Cricket
		}
		return 6 * 60 * 60 * 1000 // 6 hours for T20 / League Cricket

	case catLower == "golf":
		return 4 * 24 * 60 * 60 * 1000 // 4 days for Golf Tournaments

	case catLower == "tennis":
		return 8 * 60 * 60 * 1000 // 8 hours for long Tennis matches

	case catLower == "football":
		return 5 * 60 * 60 * 1000 // 5 hours for Football (covers extra time, penalties, delays)

	case catLower == "motorsport", catLower == "combat sports", catLower == "basketball", catLower == "baseball", catLower == "rugby":
		return 6 * 60 * 60 * 1000 // 6 hours

	default:
		return 8 * 60 * 60 * 1000 // 8 hours default max duration
	}
}
