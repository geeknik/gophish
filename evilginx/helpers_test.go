package evilginx

import (
	"net/url"
	"testing"
)

func TestGenRandomString(t *testing.T) {
	s := GenRandomString(10)
	if len(s) != 10 {
		t.Errorf("GenRandomString(10) length = %d, want 10", len(s))
	}

	s2 := GenRandomString(10)
	if s == s2 {
		t.Errorf("GenRandomString should produce different results")
	}
}

func TestGenRandomStringZeroLength(t *testing.T) {
	s := GenRandomString(0)
	if len(s) != 0 {
		t.Errorf("GenRandomString(0) length = %d, want 0", len(s))
	}
}

func TestGenRandomAlphanumString(t *testing.T) {
	s := GenRandomAlphanumString(10)
	if len(s) != 10 {
		t.Errorf("GenRandomAlphanumString(10) length = %d, want 10", len(s))
	}

	for _, c := range s {
		isLetter := (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
		isDigit := c >= '0' && c <= '9'
		if !isLetter && !isDigit {
			t.Errorf("GenRandomAlphanumString contains invalid character: %c", c)
		}
	}
}

func TestGenRandomAlphanumStringZeroLength(t *testing.T) {
	s := GenRandomAlphanumString(0)
	if len(s) != 0 {
		t.Errorf("GenRandomAlphanumString(0) length = %d, want 0", len(s))
	}
}

func TestAddPhishUrlParamsEmptyParams(t *testing.T) {
	baseURL, _ := url.Parse("https://example.com/path")
	params := url.Values{}

	AddPhishUrlParams(baseURL, params, "")

	if baseURL.RawQuery != "" {
		t.Errorf("Empty params should not add query string, got %s", baseURL.RawQuery)
	}
}

func TestAddPhishUrlParamsWithParams(t *testing.T) {
	baseURL, _ := url.Parse("https://example.com/path")
	params := url.Values{}
	params.Set("rid", "ABC1234")

	AddPhishUrlParams(baseURL, params, "")

	if baseURL.RawQuery == "" {
		t.Errorf("Params should add query string")
	}
}

func TestAddPhishUrlParamsWithBaseKey(t *testing.T) {
	baseURL, _ := url.Parse("https://example.com/path")
	params := url.Values{}
	params.Set("rid", "ABC1234")

	AddPhishUrlParams(baseURL, params, "secret-key")

	if baseURL.RawQuery == "" {
		t.Errorf("Params with base key should add query string")
	}
}
