package ingest

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAnonymizerHashing(t *testing.T) {
	anon := NewAnonymizer("test_secret_key_123")

	id1 := anon.HashVisitorID("192.168.1.1", "Mozilla/5.0", "default")
	id2 := anon.HashVisitorID("192.168.1.1", "Mozilla/5.0", "default")
	id3 := anon.HashVisitorID("192.168.1.2", "Mozilla/5.0", "default")

	if id1 == "" {
		t.Fatal("Expected non-empty visitor hash")
	}
	if id1 != id2 {
		t.Errorf("Expected identical hash for same IP and UA on same day, got %s vs %s", id1, id2)
	}
	if id1 == id3 {
		t.Errorf("Expected different hashes for different IPs, got same: %s", id1)
	}
}

func TestParseRequest(t *testing.T) {
	anon := NewAnonymizer("test_secret")

	req := httptest.NewRequest(http.MethodGet, "/e", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("CF-IPCountry", "RO")
	req.Header.Set("CF-Connecting-IP", "1.2.3.4")

	info := anon.ParseRequest(req, "https://news.ycombinator.com/item?id=123")

	if info.Browser != "Chrome" {
		t.Errorf("Expected Chrome, got %s", info.Browser)
	}
	if info.OS != "macOS" {
		t.Errorf("Expected macOS, got %s", info.OS)
	}
	if info.Device != "desktop" {
		t.Errorf("Expected desktop, got %s", info.Device)
	}
	if info.Country != "RO" {
		t.Errorf("Expected RO, got %s", info.Country)
	}
	if info.ReferrerDomain != "news.ycombinator.com" {
		t.Errorf("Expected news.ycombinator.com, got %s", info.ReferrerDomain)
	}
}
