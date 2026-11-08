package gotest

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Serves the test files of the container, so that clamav-rest (running in the same container) can download them.
// The test image sets SCAN_URL_ALLOW_PRIVATE_NETWORKS=true, otherwise downloads from localhost would be refused.
func startFileServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.FileServer(http.Dir("/clamav/tmp")))
	t.Cleanup(srv.Close)
	return srv
}

// Test scanUrl with OK file, should return 200
func TestScanURL_NonVirus(t *testing.T) {
	srv := startFileServer(t)
	got := setupScanURL(t, map[string]string{"url": srv.URL + "/ok/test.txt?X-Amz-Signature=abc"})
	if got.StatusCode != http.StatusOK {
		t.Errorf("TestScanURL_NonVirus failed, wanted %d, got %d", http.StatusOK, got.StatusCode)
	}
}

// Test scanUrl with eicar virus test file, should return 406
func TestScanURL_WithVirus(t *testing.T) {
	srv := startFileServer(t)
	got := setupScanURL(t, map[string]string{"url": srv.URL + "/virus/eicar.test"})
	if got.StatusCode != http.StatusNotAcceptable {
		t.Errorf("TestScanURL_WithVirus failed, wanted %d, got %d", http.StatusNotAcceptable, got.StatusCode)
	}
}

// Test scanUrl with a url the remote server answers with 404, should return 502
func TestScanURL_NotFound(t *testing.T) {
	srv := startFileServer(t)
	got := setupScanURL(t, map[string]string{"url": srv.URL + "/does/not/exist.txt"})
	if got.StatusCode != http.StatusBadGateway {
		t.Errorf("TestScanURL_NotFound failed, wanted %d, got %d", http.StatusBadGateway, got.StatusCode)
	}
}

// Test scanUrl without url, should return 400
func TestScanURL_MissingURL(t *testing.T) {
	got := setupScanURL(t, map[string]string{})
	if got.StatusCode != http.StatusBadRequest {
		t.Errorf("TestScanURL_MissingURL failed, wanted %d, got %d", http.StatusBadRequest, got.StatusCode)
	}
}

// Test scanUrl with an unsupported scheme, should return 403
func TestScanURL_UnsupportedScheme(t *testing.T) {
	got := setupScanURL(t, map[string]string{"url": "file:///clamav/tmp/virus/eicar.test"})
	if got.StatusCode != http.StatusForbidden {
		t.Errorf("TestScanURL_UnsupportedScheme failed, wanted %d, got %d", http.StatusForbidden, got.StatusCode)
	}
}

// setup and make call
func setupScanURL(t *testing.T, body map[string]string) *http.Response {
	t.Helper()
	reqURL, err := getURL(nil, "scanUrl")
	if err != nil {
		t.Fatalf("unable to create url, %v", err)
	}
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("unable to marshal request body, %v", err)
	}
	resp, err := c.Post(reqURL.String(), "application/json", bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("failed when sending request to clamav-rest, %v", err)
	}
	t.Cleanup(func() { resp.Body.Close() })
	return resp
}
