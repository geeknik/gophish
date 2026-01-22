package imap

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type OAuthToken struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	ExpiresAt   time.Time
}

type OAuthClient struct {
	TenantID     string
	ClientID     string
	ClientSecret string
	token        *OAuthToken
	mu           sync.Mutex
}

func NewOAuthClient(tenantID, clientID, clientSecret string) *OAuthClient {
	return &OAuthClient{
		TenantID:     tenantID,
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}
}

func (c *OAuthClient) GetAccessToken() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.token != nil && time.Now().Before(c.token.ExpiresAt.Add(-60*time.Second)) {
		return c.token.AccessToken, nil
	}

	token, err := c.fetchToken()
	if err != nil {
		return "", err
	}
	c.token = token
	return token.AccessToken, nil
}

func (c *OAuthClient) fetchToken() (*OAuthToken, error) {
	tokenURL := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", c.TenantID)

	data := url.Values{}
	data.Set("client_id", c.ClientID)
	data.Set("client_secret", c.ClientSecret)
	data.Set("scope", "https://outlook.office365.com/.default")
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to request token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var token OAuthToken
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	token.ExpiresAt = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
	return &token, nil
}
