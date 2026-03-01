package gotest

import (
	"net/http"
	"testing"
)

// Test scanPathHandler, should return 200
func TestScanPathHandler_NonVirus(t *testing.T) {
	want := http.StatusOK
	qParams := make(map[string]string)
	qParams["path"] = "/clamav/tmp/ok"
	url, err := getURL(&qParams, "scanPath")
	if err != nil {
		t.Fatalf("TestScanPathHandler_NonVirus failed when creating url, %v", err)
	}
	res, err := c.Get(url.String())
	if err != nil {
		t.Fatalf("TestScanPathHandler_NonVirus failed, wanted %d, got err: %v", want, err)
	}
	got := res.StatusCode
	if got != want {
		t.Fatalf("TestScanPathHandler_NonVirus failed, wanted %d, got %v", want, got)
	}
}

// Test scanPathHandler, should return 406
func TestScanPathHandler_WithVirus(t *testing.T) {
	want := http.StatusNotAcceptable
	qParams := make(map[string]string)
	qParams["path"] = "/clamav/tmp/virus"
	url, err := getURL(&qParams, "scanPath")
	if err != nil {
		t.Fatalf("TestScanPathHandler_WithVirus failed when creating url, %v", err)
	}
	res, err := c.Get(url.String())
	if err != nil {
		t.Fatalf("TestScanPathHandler_WithVirus failed, wanted %d, got err: %v", want, err)
	}
	got := res.StatusCode
	if got != want {
		t.Fatalf("TestScanPathHandler_NonVirus failed, wanted %d, got %v", want, got)
	}
}
