package imap

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestNewOAuthClient(t *testing.T) {
	client := NewOAuthClient("tenant-id", "client-id", "client-secret")

	if client.TenantID != "tenant-id" {
		t.Errorf("TenantID = %s, want tenant-id", client.TenantID)
	}
	if client.ClientID != "client-id" {
		t.Errorf("ClientID = %s, want client-id", client.ClientID)
	}
	if client.ClientSecret != "client-secret" {
		t.Errorf("ClientSecret = %s, want client-secret", client.ClientSecret)
	}
	if client.token != nil {
		t.Errorf("token should be nil initially")
	}
}

func TestGetAccessTokenCaching(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		resp := OAuthToken{
			AccessToken: "test-token",
			TokenType:   "Bearer",
			ExpiresIn:   3600,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &OAuthClient{
		TenantID:     "test",
		ClientID:     "test",
		ClientSecret: "test",
	}

	client.token = &OAuthToken{
		AccessToken: "cached-token",
		TokenType:   "Bearer",
		ExpiresIn:   3600,
		ExpiresAt:   time.Now().Add(30 * time.Minute),
	}

	token, err := client.GetAccessToken()
	if err != nil {
		t.Fatalf("GetAccessToken failed: %v", err)
	}

	if token != "cached-token" {
		t.Errorf("Should return cached token, got %s", token)
	}

	if callCount != 0 {
		t.Errorf("Should not call token endpoint when cache valid, called %d times", callCount)
	}
}

func TestGetAccessTokenExpiredNeedsRefresh(t *testing.T) {
	client := &OAuthClient{
		TenantID:     "test",
		ClientID:     "test",
		ClientSecret: "test",
		token: &OAuthToken{
			AccessToken: "expired-token",
			ExpiresAt:   time.Now().Add(-1 * time.Hour),
		},
	}

	client.mu.Lock()
	needsRefresh := client.token == nil || !time.Now().Before(client.token.ExpiresAt.Add(-60*time.Second))
	client.mu.Unlock()

	if !needsRefresh {
		t.Errorf("Expired token should trigger refresh")
	}
}

func TestGetAccessTokenConcurrent(t *testing.T) {
	client := &OAuthClient{
		TenantID:     "test",
		ClientID:     "test",
		ClientSecret: "test",
		token: &OAuthToken{
			AccessToken: "concurrent-token",
			ExpiresAt:   time.Now().Add(1 * time.Hour),
		},
	}

	var wg sync.WaitGroup
	results := make([]string, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			token, err := client.GetAccessToken()
			if err != nil {
				t.Errorf("GetAccessToken failed: %v", err)
				return
			}
			results[idx] = token
		}(i)
	}

	wg.Wait()

	for i, result := range results {
		if result != "concurrent-token" {
			t.Errorf("Concurrent access %d got wrong token: %s", i, result)
		}
	}
}

func TestOAuthTokenExpiration(t *testing.T) {
	token := &OAuthToken{
		AccessToken: "test",
		ExpiresIn:   3600,
		ExpiresAt:   time.Now().Add(3600 * time.Second),
	}

	if time.Now().After(token.ExpiresAt) {
		t.Errorf("Token should not be expired")
	}

	expiredToken := &OAuthToken{
		AccessToken: "expired",
		ExpiresAt:   time.Now().Add(-1 * time.Second),
	}

	if time.Now().Before(expiredToken.ExpiresAt) {
		t.Errorf("Token should be expired")
	}
}

func TestOAuthTokenExpirationBuffer(t *testing.T) {
	client := &OAuthClient{
		TenantID:     "test",
		ClientID:     "test",
		ClientSecret: "test",
		token: &OAuthToken{
			AccessToken: "about-to-expire",
			ExpiresAt:   time.Now().Add(30 * time.Second),
		},
	}

	client.mu.Lock()
	needsRefresh := client.token == nil || !time.Now().Before(client.token.ExpiresAt.Add(-60*time.Second))
	client.mu.Unlock()

	if !needsRefresh {
		t.Errorf("Token expiring within 60 seconds should trigger refresh")
	}
}
