package gotest

import (
	"net/http"
	"testing"
)

// setup and make call
func setupScanPathHandler(path string, t *testing.T, want int) *http.Response {
	qParams := make(map[string]string, 1)
	qParams["path"] = path
	url, err := getURL(&qParams, "scanPath")
	if err != nil {
		t.Fatalf("TestScanPathHandler_NonVirus failed when creating url, %v", err)
	}
	res, err := c.Get(url.String())
	if err != nil {
		t.Errorf("TestScanPathHandler_NonVirus failed, wanted %v, got err: %v", want, err)
	}
	return res
}

// Test scanPathHandler, should return 200
func TestScanPathHandler_NonVirus(t *testing.T) {
	want := http.StatusOK
	path := "/clamav/tmp/ok"
	res := setupScanPathHandler(path, t, want)
	got := res.StatusCode
	if got != want {
		t.Errorf("TestScanPathHandler_NonVirus failed, wanted %d, got %v", want, got)
	}
}

// Test scanPathHandler, should return 406
func TestScanPathHandler_WithVirus(t *testing.T) {
	want := http.StatusNotAcceptable
	path := "/clamav/tmp/virus"
	res := setupScanPathHandler(path, t, want)
	got := res.StatusCode
	if got != want {
		t.Errorf("TestScanPathHandler_NonVirus failed, wanted %d, got %v", want, got)
	}
}
