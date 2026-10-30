package ingest

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// Anonymizer generates privacy-compliant daily visitor IDs without storing IP addresses.
type Anonymizer struct {
	secretKey  string
	currentDay string
	dailySalt  string
	mu         sync.RWMutex
}

// NewAnonymizer creates an Anonymizer configured with a secret key.
func NewAnonymizer(secretKey string) *Anonymizer {
	a := &Anonymizer{
		secretKey: secretKey,
	}
	a.refreshDailySalt()
	return a
}

func (a *Anonymizer) refreshDailySalt() {
	today := time.Now().UTC().Format("2006-01-02")
	if a.currentDay != today {
		h := sha256.Sum256([]byte(a.secretKey + today))
		a.currentDay = today
		a.dailySalt = hex.EncodeToString(h[:16])
	}
}

// HashVisitorID produces an anonymous, irreversible visitor identifier that changes every 24 hours.
func (a *Anonymizer) HashVisitorID(ip, userAgent, siteID string) string {
	a.mu.Lock()
	a.refreshDailySalt()
	salt := a.dailySalt
	a.mu.Unlock()

	raw := ip + "|" + userAgent + "|" + siteID + "|" + salt
	hash := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hash[:16])
}

// ClientInfo holds parsed device and location details from the HTTP request.
type ClientInfo struct {
	IP             string
	UserAgent      string
	Browser        string
	OS             string
	Device         string
	Country        string
	ReferrerDomain string
}

// ParseRequest extracts client environment metadata without persisting raw identifiers.
func (a *Anonymizer) ParseRequest(r *http.Request, rawReferrer string) ClientInfo {
	ip := extractClientIP(r)
	ua := r.UserAgent()

	browser := detectBrowser(ua)
	osName := detectOS(ua)
	device := detectDevice(ua)
	country := extractCountry(r)
	refDomain := extractDomain(rawReferrer)

	return ClientInfo{
		IP:             ip,
		UserAgent:      ua,
		Browser:        browser,
		OS:             osName,
		Device:         device,
		Country:        country,
		ReferrerDomain: refDomain,
	}
}

func extractClientIP(r *http.Request) string {
	// Look for standard reverse proxy headers
	headers := []string{"CF-Connecting-IP", "X-Real-IP", "X-Forwarded-For"}
	for _, h := range headers {
		val := r.Header.Get(h)
		if val != "" {
			parts := strings.Split(val, ",")
			return strings.TrimSpace(parts[0])
		}
	}
	// Fallback to RemoteAddr
	parts := strings.Split(r.RemoteAddr, ":")
	if len(parts) > 0 {
		return parts[0]
	}
	return r.RemoteAddr
}

func extractCountry(r *http.Request) string {
	// Common CDN GeoIP headers
	headers := []string{
		"CF-IPCountry",
		"X-Country-Code",
		"CloudFront-Viewer-Country",
		"X-GeoIP-Country",
	}
	for _, h := range headers {
		code := strings.TrimSpace(r.Header.Get(h))
		if code != "" && len(code) == 2 && code != "XX" && code != "T1" {
			return strings.ToUpper(code)
		}
	}
	return "US"
}

func detectBrowser(ua string) string {
	lower := strings.ToLower(ua)
	switch {
	case strings.Contains(lower, "edg/"):
		return "Edge"
	case strings.Contains(lower, "opr/") || strings.Contains(lower, "opera"):
		return "Opera"
	case strings.Contains(lower, "chrome") || strings.Contains(lower, "crios"):
		return "Chrome"
	case strings.Contains(lower, "firefox") || strings.Contains(lower, "fxios"):
		return "Firefox"
	case strings.Contains(lower, "safari") && !strings.Contains(lower, "chrome"):
		return "Safari"
	case strings.Contains(lower, "samsungbrowser"):
		return "Samsung Internet"
	default:
		return "Other"
	}
}

func detectOS(ua string) string {
	lower := strings.ToLower(ua)
	switch {
	case strings.Contains(lower, "windows"):
		return "Windows"
	case strings.Contains(lower, "macintosh") || strings.Contains(lower, "mac os"):
		return "macOS"
	case strings.Contains(lower, "android"):
		return "Android"
	case strings.Contains(lower, "iphone") || strings.Contains(lower, "ipad") || strings.Contains(lower, "ipod"):
		return "iOS"
	case strings.Contains(lower, "linux"):
		return "Linux"
	case strings.Contains(lower, "cros"):
		return "ChromeOS"
	default:
		return "Other"
	}
}

func detectDevice(ua string) string {
	lower := strings.ToLower(ua)
	switch {
	case strings.Contains(lower, "ipad") || strings.Contains(lower, "tablet"):
		return "tablet"
	case strings.Contains(lower, "mobile") || strings.Contains(lower, "iphone") || strings.Contains(lower, "android"):
		return "mobile"
	default:
		return "desktop"
	}
}

func extractDomain(rawURL string) string {
	if rawURL == "" {
		return "direct"
	}
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Host == "" {
		return "direct"
	}
	host := parsed.Host
	if colonIdx := strings.Index(host, ":"); colonIdx != -1 {
		host = host[:colonIdx]
	}
	return strings.TrimPrefix(host, "www.")
}
