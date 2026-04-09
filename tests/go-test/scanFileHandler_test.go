package gotest

import (
	"net/http"
	"testing"
)

// Test scanFileHandler, should return 200
func TestScanFileHandler_NonVirus(t *testing.T) {
	want := http.StatusOK
	path := "/clamav/tmp/ok/test.txt"
	res := setupScanFileHandler(path, t, want)
	got := res.StatusCode
	if got != want {
		t.Errorf("ScanFileTestScanFileHandler_NonVirus failed, wanted %d, got %d", want, got)
	}
}

// Test scanFileHandler, should return 406
func TestScanFileHandler_WithVirus(t *testing.T) {
	want := http.StatusNotAcceptable
	path := "/clamav/tmp/virus/eicar.test"
	res := setupScanFileHandler(path, t, want)
	got := res.StatusCode
	if got != want {
		t.Errorf("TestScanFileHandler_WithVirus failed, wanted %d, got %v", want, got)
	}
}

// Setup and make call
func setupScanFileHandler(path string, t *testing.T, want int) *http.Response {
	qParams := make(map[string]string, 1)
	qParams["path"] = path
	url, err := getURL(&qParams, "scanFile")
	if err != nil {
		t.Fatalf("TestScanFileHandler_NonVirus failed when creating url, %v", err)
	}
	res, err := c.Get(url.String())
	if err != nil {
		t.Fatalf("TestScanFileHandler_NonVirus failed, wanted %d, got err: %v", want, err)
	}
	return res
}
