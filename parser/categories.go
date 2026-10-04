package parser

import (
	"strings"
	"github.com/neelniloy/lstv-event-parser/models"
)

func isAlphaNum(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')
}

// containsWord checks if word appears as an isolated token in s (bounded by non-alphanumeric chars or string ends)
func containsWord(s, word string) bool {
	if word == "" || s == "" {
		return false
	}
	idx := strings.Index(s, word)
	for idx != -1 {
		startOk := idx == 0 || !isAlphaNum(s[idx-1])
		endIdx := idx + len(word)
		endOk := endIdx == len(s) || !isAlphaNum(s[endIdx])
		if startOk && endOk {
			return true
		}
		next := strings.Index(s[idx+1:], word)
		if next == -1 {
			break
		}
		idx += 1 + next
	}
	return false
}

func GetSportCategory(event *models.TimelineEvent) string {
	titleLower := strings.ToLower(event.Title)
	catLower := strings.TrimSpace(strings.ToLower(event.Category))
	leagueLower := strings.ToLower(event.League)

	// Infer league if missing
	if event.League == "" {
		event.League = InferLeague(titleLower, catLower)
		leagueLower = strings.ToLower(event.League)
	}

	combined := titleLower + " | " + leagueLower + " | " + strings.ToLower(event.Description)

	// 1. Motorsport
	motorsportKeywords := []string{
		"formula 1", "formula1", "formula one", "motogp", "nascar", "indycar",
		"rally", "motorsport", "off-road", "off road", "supercars", "dakar",
		"grand prix", "speedway",
	}
	for _, kw := range motorsportKeywords {
		if strings.Contains(combined, kw) || strings.Contains(catLower, kw) {
			return "Motorsport"
		}
	}
	if containsWord(combined, "f1") || containsWord(catLower, "f1") {
		return "Motorsport"
	}

	// 2. Combat Sports
	combatKeywords := []string{
		"ufc", "mma", "wwe", "wrestling", "boxing", "one championship", "one friday fights", "bellator",
		"pfl", "bare knuckle", "bkfc", "aew", "smackdown", "raw", "nxt", "dynamite", "ring of combat",
		"combate global", "glory kickboxing", "muay thai", "judo", "karate", "taekwondo", "bjj", "jiu jitsu",
	}
	for _, kw := range combatKeywords {
		if strings.Contains(combined, kw) || strings.Contains(catLower, kw) {
			return "Combat Sports"
		}
	}

	// 3. Tennis & Racket Sports
	tennisKeywords := []string{
		"tennis", "wimbledon", "roland garros", "french open", "australian open", "us open tennis",
		"us open", "padel", "world padel", "premier padel", "pickleball", "grand slam",
		"cincinnati open", "madrid open", "italian open", "internazionali bnl", "miami open", "indian wells",
		"monte-carlo", "monte carlo", "shanghai masters", "paris masters", "canadian open", "national bank open",
		"china open", "japan open", "beijing open", "tokyo open", "barcelona open", "rotterdam open", "rio open",
		"dubai tennis", "qatar open", "acapulco open", "halle open", "queen's club", "atp finals", "wta finals",
		"davis cup", "billie jean king cup", "fed cup", "laver cup", "united cup", "hopman cup",
	}
	for _, kw := range tennisKeywords {
		if strings.Contains(combined, kw) || strings.Contains(catLower, kw) {
			return "Tennis"
		}
	}
	if containsWord(combined, "atp") || containsWord(combined, "wta") || containsWord(combined, "itf") {
		return "Tennis"
	}

	// 4. Basketball
	basketballKeywords := []string{"basketball", "euroleague", "fiba", "ncaa basketball", "march madness"}
	for _, kw := range basketballKeywords {
		if strings.Contains(combined, kw) || strings.Contains(catLower, kw) {
			return "Basketball"
		}
	}
	if containsWord(combined, "nba") || containsWord(combined, "wnba") || containsWord(catLower, "nba") || containsWord(catLower, "wnba") {
		return "Basketball"
	}

	// 5. Golf
	golfKeywords := []string{
		"golf", "pga tour", "dp world tour", "liv golf", "masters tournament", "ryder cup", "lpga",
		"us open golf", "the open championship", "tourcast", "fedexcup", "fedex st. jude",
	}
	for _, kw := range golfKeywords {
		if strings.Contains(combined, kw) || strings.Contains(catLower, kw) {
			return "Golf"
		}
	}

	// 6. Baseball
	baseballKeywords := []string{"baseball", "world baseball classic"}
	for _, kw := range baseballKeywords {
		if strings.Contains(combined, kw) || strings.Contains(catLower, kw) {
			return "Baseball"
		}
	}
	if containsWord(combined, "mlb") || containsWord(combined, "kbo") || containsWord(combined, "npb") || containsWord(catLower, "mlb") {
		return "Baseball"
	}

	// 7. Ice Hockey & Field Hockey
	if containsWord(combined, "nhl") || containsWord(combined, "ahl") || containsWord(combined, "khl") ||
		containsWord(combined, "shl") || containsWord(combined, "iihf") ||
		strings.Contains(combined, "ice hockey") || strings.Contains(combined, "stanley cup") {
		return "Ice Hockey"
	}
	if strings.Contains(combined, "fih hockey") || strings.Contains(combined, "field hockey") || strings.Contains(combined, "fih") {
		return "Hockey"
	}

	// 8. Rugby & American Football
	rugbyKeywords := []string{"rugby", "six nations", "nrl", "super rugby", "top 14", "premiership rugby", "united rugby championship"}
	for _, kw := range rugbyKeywords {
		if strings.Contains(combined, kw) || strings.Contains(catLower, kw) {
			return "Rugby"
		}
	}

	amFootballKeywords := []string{"super bowl", "ncaa football", "college football", "american football"}
	for _, kw := range amFootballKeywords {
		if strings.Contains(combined, kw) || strings.Contains(catLower, kw) {
			return "American Football"
		}
	}
	if containsWord(combined, "nfl") || containsWord(combined, "cfl") || containsWord(catLower, "nfl") {
		return "American Football"
	}

	// 9. Badminton, Volleyball, Esports
	if strings.Contains(combined, "badminton") || containsWord(combined, "bwf") {
		return "Badminton"
	}
	if strings.Contains(combined, "volleyball") || containsWord(combined, "vnl") || containsWord(combined, "fivb") {
		return "Volleyball"
	}
	if strings.Contains(combined, "esports") || strings.Contains(combined, "valorant") ||
		strings.Contains(combined, "cs:go") || strings.Contains(combined, "cs2") ||
		strings.Contains(combined, "dota 2") || strings.Contains(combined, "league of legends") {
		return "Esports"
	}

	// 10. Distinct Cricket Tournaments, Leagues & Cues
	cricketDistinct := []string{
		"cricket", "t20", "odi", "t10", "ipl", "indian premier league", "bpl", "bangladesh premier league",
		"psl", "pakistan super league", "cpl t20", "cpl", "caribbean premier league", "big bash", "bbl", "wbbl",
		"the hundred", "vitality blast", "t20 blast", "county championship", "ranji trophy", "super smash",
		"major league cricket", "mlc", "world championship of legends", "wcl", "asia cup", "tri-series",
		"ashes", "test match", "test series", "1st test", "2nd test", "3rd test", "4th test", "5th test",
		"super giants", "trent rockets", "london spirit", "manchester originals", "welsh fire", "southern brave",
	}
	for _, kw := range cricketDistinct {
		if strings.Contains(combined, kw) || containsWord(combined, kw) {
			return "Cricket"
		}
	}

	// Cricket tours involving major cricket playing nations
	cricketNations := []string{"india", "pakistan", "bangladesh", "england", "australia", "new zealand", "south africa", "west indies", "sri lanka", "afghanistan", "ireland", "zimbabwe"}
	if strings.Contains(combined, "tour of") {
		for _, team := range cricketNations {
			if strings.Contains(combined, team) {
				return "Cricket"
			}
		}
	}

	// 11. Football (Soccer) - leagues, tournaments, teams, clubs
	footballKeywords := []string{
		"football", "soccer", "calcio", "futbol", "futebol",
		// Leagues & Tournaments
		"premier league", "la liga", "laliga", "segunda división", "segunda division", "copa del rey",
		"serie a", "serie b", "serie c", "serie d", "coppa italia", "supercoppa",
		"bundesliga", "2. bundesliga", "dfb-pokal", "dfb pokal",
		"ligue 1", "ligue 2", "coupe de france", "trophée des champions",
		"champions league", "europa league", "conference league", "uefa",
		"major league soccer", "nwsl", "usl",
		"copa libertadores", "copa sudamericana", "recopa",
		"eredivisie", "primeira liga", "segunda liga", "liga portugal", "taça de portugal",
		"süper lig", "super lig", "scottish premiership", "spfl", "jupiler pro league",
		"liga mx", "liga bbva mx", "liga profesional", "primera división", "primera division",
		"division profesional", "liga i", "hnl", "nb i", "first league", "i liga", "3. liga",
		"efl", "championship", "fa cup", "carabao cup", "friendlies clubs", "club friendlies",
		"copa america", "fifa", "nations league", "concacaf", "afcon", "africa cup of nations",
		// Top clubs & teams
		"real madrid", "barcelona", "bayern", "juventus", "psg", "chelsea", "manchester united", "manchester city",
		"man utd", "man city", "liverpool", "arsenal", "tottenham", "ac milan", "inter milan", "atletico madrid",
		"atletico-mg", "boca juniors", "river plate", "flamengo", "palmeiras", "corinthians", "sao paulo", "gremio",
		"santos", "porto", "benfica", "sporting cp", "ajax", "feyenoord", "psv", "dortmund", "borussia", "leverkusen",
		"leipzig", "roma", "lazio", "napoli", "fiorentina", "atalanta", "newcastle", "aston villa", "west ham",
		"everton", "wolves", "sevilla", "villarreal", "athletic bilbao", "real sociedad", "betis", "valencia",
		"celta vigo", "osasuna", "las palmas", "mallorca", "getafe", "rayo vallecano", "alaves", "espanyol",
		"valladolid", "leganes", "monaco", "marseille", "lyon", "lille", "nice", "lens", "rennes",
		"galatasaray", "fenerbahce", "besiktas", "al-nassr", "al-hilal", "al-ittihad", "inter miami", "la galaxy",
		"como 1907", "calcio como", "hellas verona", "virtus entella", "sporting gijon", "albacete", "sc braga",
		"gil vicente", "famalicao", "maritimo", "club america", "atletico san luis", "santos laguna", "chivas",
		"cruz azul", "tijuana", "austin vs", "fc dallas", "seattle sounders", "vancouver whitecaps", "portland timbers",
		"chicago fire", "new york city fc", "philadelphia union", "mirassol", "botafogo", "vitoria", "cruzeiro",
		"sporting cristal", "frosinone", "juve stabia", "başakşehir", "kocaelispor", "varazdin", "rijeka", "genoa",
		"ascoli", "girona", "cordoba", "arouca", "moreirense", "beşiktaş", "eyüpspor", "corvinul", "cluj",
		"erzurumspor", "hajduk split", "caxias", "figueirense", "cobresal", "anapolis", "floresta", "mantova",
		"universitario", "olimpia", "ameliano", "confiança", "maringá", "novorizontino", "colo colo",
		"houston dash", "washington spirit", "rosario central", "cienciano", "garcilaso", "ferencvarosi",
		"empoli", "sassuolo", "cesena", "spartak kostroma",
	}
	for _, kw := range footballKeywords {
		if strings.Contains(combined, kw) || strings.Contains(catLower, kw) {
			return "Football"
		}
	}
	if containsWord(combined, "epl") || containsWord(combined, "mls") || containsWord(combined, "ucl") || containsWord(combined, "uel") {
		return "Football"
	}

	// Club suffix cues like " FC", " CF", " SC", " United", " City" in match title
	if (strings.Contains(combined, " fc") || strings.Contains(combined, "fc ") ||
		strings.Contains(combined, " cf") || strings.Contains(combined, "cf ") ||
		strings.Contains(combined, " sc") || strings.Contains(combined, "sc ") ||
		strings.Contains(combined, " united") || strings.Contains(combined, " city")) &&
		(strings.Contains(combined, "vs") || strings.Contains(combined, " v ") || strings.Contains(combined, " at ")) {
		return "Football"
	}

	// 12. World Cup / Tournament Context
	if strings.Contains(combined, "world cup") {
		if strings.Contains(combined, "fifa") || strings.Contains(combined, "soccer") || strings.Contains(combined, "football") {
			return "Football"
		}
		if strings.Contains(combined, "cricket") || strings.Contains(combined, "t20") || strings.Contains(combined, "icc") {
			return "Cricket"
		}
		if strings.Contains(combined, "rugby") {
			return "Rugby"
		}
		if strings.Contains(combined, "hockey") || strings.Contains(combined, "fih") {
			return "Hockey"
		}
		for _, team := range cricketNations {
			if strings.Contains(combined, team) {
				return "Cricket"
			}
		}
		return "Football"
	}

	// 13. Fallback to source category if valid and specific
	if strings.Contains(catLower, "cricket") {
		return "Cricket"
	}
	if strings.Contains(catLower, "football") || strings.Contains(catLower, "soccer") {
		return "Football"
	}
	if strings.Contains(catLower, "tennis") {
		return "Tennis"
	}
	if strings.Contains(catLower, "hockey") {
		return "Ice Hockey"
	}

	if event.Category != "" && !strings.EqualFold(event.Category, "Sports") &&
		!strings.EqualFold(event.Category, "Other") && !strings.EqualFold(event.Category, "Other Sports") &&
		!strings.EqualFold(event.Category, "General") {
		return strings.TrimSpace(event.Category)
	}

	return "Other Sports"
}

func InferLeague(titleLower, catLower string) string {
	switch {
	case strings.Contains(titleLower, "cpl") || strings.Contains(titleLower, "caribbean premier league"):
		return "Caribbean Premier League"
	case strings.Contains(titleLower, "ipl") || strings.Contains(titleLower, "indian premier league"):
		return "Indian Premier League"
	case strings.Contains(titleLower, "bpl") || strings.Contains(titleLower, "bangladesh premier league"):
		return "Bangladesh Premier League"
	case strings.Contains(titleLower, "psl") || strings.Contains(titleLower, "pakistan super league"):
		return "Pakistan Super League"
	case (strings.Contains(titleLower, "premier league") && !strings.Contains(titleLower, "caribbean") && !strings.Contains(titleLower, "indian") && !strings.Contains(titleLower, "bangladesh") && !strings.Contains(titleLower, "pakistan")) || strings.Contains(titleLower, "epl"):
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
