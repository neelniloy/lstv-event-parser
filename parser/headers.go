package parser

import (
	"html"
	"strings"
)

func UnescapeHTML(text string) string {
	if strings.TrimSpace(text) == "" {
		return text
	}
	return html.UnescapeString(text)
}

func SanitizeStreamURL(rawURL string) string {
	if strings.TrimSpace(rawURL) == "" {
		return rawURL
	}
	cleaned := strings.TrimSpace(rawURL)
	if strings.Contains(cleaned, "bd-mc-pdlive.fancode.com") {
		cleaned = strings.ReplaceAll(cleaned, "bd-mc-pdlive.fancode.com", "dai-fancode.pages.dev")
	}
	return cleaned
}

func IsStreamPlayable(rawURL string) bool {
	u := SanitizeStreamURL(rawURL)
	if u == "" || !strings.HasPrefix(u, "http") {
		return false
	}
	lower := strings.ToLower(u)
	return strings.Contains(lower, ".m3u8") ||
		strings.Contains(lower, ".mpd") ||
		strings.Contains(lower, ".mp4") ||
		strings.Contains(lower, ".ts") ||
		strings.Contains(lower, "playerado.top") ||
		strings.Contains(lower, "cricstreams.org")
}

func GetStreamHeadersForURL(streamURL string, existingHeaders map[string]string) map[string]string {
	headers := make(map[string]string)
	for k, v := range existingHeaders {
		headers[k] = v
	}

	if streamURL == "" || !strings.HasPrefix(streamURL, "http") {
		return headers
	}

	lower := strings.ToLower(streamURL)
	hasReferer := false
	hasOrigin := false

	for k := range headers {
		if strings.EqualFold(k, "Referer") {
			hasReferer = true
		}
		if strings.EqualFold(k, "Origin") {
			hasOrigin = true
		}
	}

	switch {
	case strings.Contains(lower, "aiv-cdn.net") || strings.Contains(lower, "amazon.com") || strings.Contains(lower, "akamaihd.net/ottb") || strings.Contains(lower, "pv-cdn.net"):
		if !hasReferer {
			headers["Referer"] = "https://www.amazon.com/"
		}
		if !hasOrigin {
			headers["Origin"] = "https://www.amazon.com"
		}
		if _, ok := headers["User-Agent"]; !ok {
			headers["User-Agent"] = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
		}

	case strings.Contains(lower, "sonyliv") || strings.Contains(lower, "slivcdn") || strings.Contains(lower, "sonyeventsglobal"):
		if !hasReferer {
			headers["Referer"] = "https://www.sonyliv.com/"
		}
		if !hasOrigin {
			headers["Origin"] = "https://www.sonyliv.com"
		}
		if _, ok := headers["User-Agent"]; !ok {
			headers["User-Agent"] = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
		}

	case strings.Contains(lower, "tapmad") || strings.Contains(lower, "saseries"):
		if !hasReferer {
			headers["Referer"] = "https://www.tapmad.com/"
		}
		if !hasOrigin {
			headers["Origin"] = "https://www.tapmad.com"
		}
		if _, ok := headers["User-Agent"]; !ok {
			headers["User-Agent"] = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
		}

	case strings.Contains(lower, "cricstreams") || strings.Contains(lower, "playerado") || strings.Contains(lower, "crichd") || strings.Contains(lower, "johirxofficial"):
		if !hasReferer {
			headers["Referer"] = "https://cricstreams.org/"
		}
		if !hasOrigin {
			headers["Origin"] = "https://cricstreams.org"
		}

	case strings.Contains(lower, "fancode") || strings.Contains(lower, "pages.dev"):
		if !hasReferer {
			headers["Referer"] = "https://www.fancode.com/"
		}
		if _, ok := headers["User-Agent"]; !ok {
			headers["User-Agent"] = "ReactNativeVideo/9.7.0 (Linux;Android 10)"
		}

	case strings.Contains(lower, "willow"):
		if !hasReferer {
			headers["Referer"] = "https://www.willow.tv/"
		}
		if !hasOrigin {
			headers["Origin"] = "https://www.willow.tv"
		}

	case strings.Contains(lower, "tmaxapp") || strings.Contains(lower, "tvdsz") || strings.Contains(lower, "live.php") || strings.Contains(lower, "mac="):
		if _, ok := headers["User-Agent"]; !ok {
			headers["User-Agent"] = "IPTVSmartersPlayer"
		}

	default:
		// Do NOT synthesize Referer/Origin for generic IPTV/HLS streams.
		// Upstream relay servers reject cross-domain referers with 400 Bad Request.
	}

	return headers
}
