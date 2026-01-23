//go:build ignore
// +build ignore

// Test script for validating Gophish email integration with Mailgun.
//
// Environment variables required:
//   MAILGUN_API_KEY     - Your Mailgun API key
//   MAILGUN_DOMAIN      - Your Mailgun domain (e.g., sandbox123.mailgun.org)
//   TEST_EMAIL          - Email address to send test to (must be authorized for sandbox)
//
// For sandbox domains, you must first authorize the recipient:
//   curl -X POST "https://api.mailgun.net/v5/sandbox/auth_recipients?email=YOUR_EMAIL" \
//     --user 'api:YOUR_API_KEY'
//
// Then verify by clicking the link in the email sent to that address.
//
// Usage:
//   MAILGUN_API_KEY=xxx MAILGUN_DOMAIN=sandbox123.mailgun.org TEST_EMAIL=you@example.com go run scripts/test_email.go

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

func main() {
	apiKey := os.Getenv("MAILGUN_API_KEY")
	domain := os.Getenv("MAILGUN_DOMAIN")
	testEmail := os.Getenv("TEST_EMAIL")

	if apiKey == "" || domain == "" {
		fmt.Println("ERROR: Required environment variables not set")
		fmt.Println("")
		fmt.Println("Usage:")
		fmt.Println("  MAILGUN_API_KEY=xxx MAILGUN_DOMAIN=your-domain.mailgun.org TEST_EMAIL=you@example.com go run scripts/test_email.go")
		fmt.Println("")
		fmt.Println("For sandbox domains, first authorize the recipient:")
		fmt.Println("  curl -X POST \"https://api.mailgun.net/v5/sandbox/auth_recipients?email=YOUR_EMAIL\" --user 'api:YOUR_API_KEY'")
		os.Exit(1)
	}

	if testEmail == "" {
		testEmail = "test@example.com"
	}

	baseURL := "https://api.mailgun.net/v3"

	fmt.Println("=== Gophish Email Integration Test ===")
	fmt.Printf("Domain: %s\n", domain)
	fmt.Printf("Test Email: %s\n\n", testEmail)

	// Test 1: API Authentication
	fmt.Println("1. Testing API authentication...")
	if !testAuth(baseURL, apiKey) {
		fmt.Println("   FAILED: API authentication failed")
		os.Exit(1)
	}
	fmt.Println("   SUCCESS: API authentication works\n")

	// Test 2: Domain Status
	fmt.Println("2. Checking domain status...")
	isSandbox := checkDomain(baseURL, apiKey, domain)
	fmt.Println()

	// Test 3: SMTP Connectivity
	fmt.Println("3. Testing SMTP connection (port 587)...")
	testSMTPConnection()
	fmt.Println()

	// Test 4: Send Email
	fmt.Println("4. Sending test email...")
	if isSandbox {
		fmt.Println("   NOTE: Sandbox domain - recipient must be authorized first")
	}
	sendTestEmail(baseURL, apiKey, domain, testEmail)
}

func testAuth(baseURL, apiKey string) bool {
	req, _ := http.NewRequest("GET", baseURL+"/domains", nil)
	req.SetBasicAuth("api", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200
}

func checkDomain(baseURL, apiKey, domain string) bool {
	req, _ := http.NewRequest("GET", baseURL+"/domains/"+domain, nil)
	req.SetBasicAuth("api", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	isSandbox := false
	if domainData, ok := result["domain"].(map[string]interface{}); ok {
		fmt.Printf("   Domain: %s\n", domainData["name"])
		fmt.Printf("   State: %s\n", domainData["state"])
		fmt.Printf("   Type: %s\n", domainData["type"])
		if domainData["type"] == "sandbox" {
			isSandbox = true
		}
	} else {
		fmt.Printf("   Response: %s\n", string(body))
	}

	return isSandbox
}

func testSMTPConnection() {
	conn, err := net.DialTimeout("tcp", "smtp.mailgun.org:587", 10*time.Second)
	if err != nil {
		fmt.Printf("   FAILED: %v\n", err)
		return
	}
	conn.Close()
	fmt.Println("   SUCCESS: SMTP port 587 is reachable")

	conn, err = net.DialTimeout("tcp", "smtp.mailgun.org:465", 10*time.Second)
	if err != nil {
		fmt.Printf("   INFO: Port 465 (SMTPS) not reachable: %v\n", err)
		return
	}
	conn.Close()
	fmt.Println("   SUCCESS: SMTP port 465 (SMTPS) is reachable")
}

func sendTestEmail(baseURL, apiKey, domain, to string) {
	data := url.Values{}
	data.Set("from", fmt.Sprintf("Gophish Test <postmaster@%s>", domain))
	data.Set("to", to)
	data.Set("subject", "Gophish Email Integration Test")
	data.Set("text", `This is a test email from Gophish.

If you receive this, the email integration is working correctly.

Test details:
- Sent via Mailgun API
- Domain: `+domain+`
- Timestamp: `+time.Now().Format(time.RFC3339))

	data.Set("html", `<!DOCTYPE html>
<html>
<head><title>Gophish Test</title></head>
<body style="font-family: Arial, sans-serif; padding: 20px;">
<h1 style="color: #333;">Gophish Email Test</h1>
<p>This is a test email from Gophish.</p>
<p>If you receive this, the email integration is working correctly.</p>
<hr>
<p style="color: #666; font-size: 12px;">
Domain: `+domain+`<br>
Timestamp: `+time.Now().Format(time.RFC3339)+`
</p>
</body>
</html>`)

	req, _ := http.NewRequest("POST",
		fmt.Sprintf("%s/%s/messages", baseURL, domain),
		strings.NewReader(data.Encode()))
	req.SetBasicAuth("api", apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if resp.StatusCode == 200 {
		fmt.Println("   SUCCESS: Email queued for delivery")
		if id, ok := result["id"].(string); ok {
			fmt.Printf("   Message ID: %s\n", id)
		}
	} else {
		fmt.Printf("   Status: %d\n", resp.StatusCode)
		fmt.Printf("   Response: %s\n", string(body))

		if strings.Contains(string(body), "not allowed to send") {
			fmt.Println("\n   HINT: For sandbox domains, authorize the recipient first:")
			fmt.Printf("   curl -X POST \"https://api.mailgun.net/v5/sandbox/auth_recipients?email=%s\" --user 'api:YOUR_API_KEY'\n", to)
		}
		if strings.Contains(string(body), "activate") {
			fmt.Println("\n   HINT: Activate your Mailgun account first (check your inbox)")
		}
	}
}
