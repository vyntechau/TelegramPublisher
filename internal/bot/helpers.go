package bot

import (
	"net/url"
	"strings"
)

// isValidPublicURL validates whether a URL string is a valid, publicly-routable HTTP/HTTPS URL.
// Telegram API rejects 'localhost', loopback IPs ('127.0.0.1', '0.0.0.0'), private/internal hostnames,
// and requires HTTPS protocol for all WebApps.
func isValidPublicURL(raw string, requireHTTPS bool) bool {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return false
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return false
	}
	scheme := strings.ToLower(u.Scheme)
	if requireHTTPS {
		if scheme != "https" {
			return false
		}
	} else {
		if scheme != "http" && scheme != "https" {
			return false
		}
	}
	host := strings.ToLower(u.Hostname())
	if host == "localhost" || host == "127.0.0.1" || host == "0.0.0.0" || strings.HasSuffix(host, ".local") || strings.HasSuffix(host, ".internal") {
		return false
	}
	if !strings.Contains(host, ".") {
		return false
	}
	return true
}
