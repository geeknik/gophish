package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gophish/gophish/models"
)

func makeImportRequest(ctx *testContext, url string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/import/site",
		bytes.NewBuffer([]byte(fmt.Sprintf(`
			{
				"url" : "%s"
			}
		`, url))))
	req.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	ctx.apiServer.ImportSite(response, req)
	return response
}

func TestSSRFProtectionBlocksMetadata(t *testing.T) {
	ctx := setupTest(t)
	metadataURL := "http://169.254.169.254/latest/meta-data/"
	response := makeImportRequest(ctx, metadataURL)
	expectedCode := http.StatusBadRequest
	if response.Code != expectedCode {
		t.Fatalf("incorrect status code received. expected %d got %d", expectedCode, response.Code)
	}
	got := &models.Response{}
	err := json.NewDecoder(response.Body).Decode(got)
	if err != nil {
		t.Fatalf("error decoding body: %v", err)
	}
	if !strings.Contains(got.Message, "upstream connection denied") {
		t.Fatalf("incorrect response error provided: %s", got.Message)
	}
}

func TestSSRFProtectionBlocksLocalhost(t *testing.T) {
	ctx := setupTest(t)
	localhostURL := "http://127.0.0.1:8080/test"
	response := makeImportRequest(ctx, localhostURL)
	expectedCode := http.StatusBadRequest
	if response.Code != expectedCode {
		t.Fatalf("incorrect status code received. expected %d got %d", expectedCode, response.Code)
	}
	got := &models.Response{}
	err := json.NewDecoder(response.Body).Decode(got)
	if err != nil {
		t.Fatalf("error decoding body: %v", err)
	}
	if !strings.Contains(got.Message, "upstream connection denied") {
		t.Fatalf("incorrect response error provided: %s", got.Message)
	}
}

func TestSSRFProtectionBlocksPrivateNetworks(t *testing.T) {
	ctx := setupTest(t)
	privateURLs := []string{
		"http://10.0.0.1/",
		"http://172.16.0.1/",
		"http://192.168.1.1/",
	}
	for _, url := range privateURLs {
		response := makeImportRequest(ctx, url)
		expectedCode := http.StatusBadRequest
		if response.Code != expectedCode {
			t.Fatalf("URL %s: incorrect status code received. expected %d got %d", url, expectedCode, response.Code)
		}
		got := &models.Response{}
		err := json.NewDecoder(response.Body).Decode(got)
		if err != nil {
			t.Fatalf("error decoding body: %v", err)
		}
		if !strings.Contains(got.Message, "upstream connection denied") {
			t.Fatalf("URL %s: incorrect response error provided: %s", url, got.Message)
		}
	}
}

func TestURLSchemeValidation(t *testing.T) {
	ctx := setupTest(t)
	invalidURLs := []struct {
		url      string
		errorMsg string
	}{
		{"file:///etc/passwd", "URL scheme must be http or https"},
		{"gopher://127.0.0.1:25/", "URL scheme must be http or https"},
		{"ftp://example.com/", "URL scheme must be http or https"},
	}
	for _, tc := range invalidURLs {
		response := makeImportRequest(ctx, tc.url)
		expectedCode := http.StatusBadRequest
		if response.Code != expectedCode {
			t.Fatalf("URL %s: incorrect status code received. expected %d got %d", tc.url, expectedCode, response.Code)
		}
		got := &models.Response{}
		err := json.NewDecoder(response.Body).Decode(got)
		if err != nil {
			t.Fatalf("error decoding body: %v", err)
		}
		if !strings.Contains(got.Message, tc.errorMsg) {
			t.Fatalf("URL %s: expected error containing '%s', got: %s", tc.url, tc.errorMsg, got.Message)
		}
	}
}
