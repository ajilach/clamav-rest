package gotest

import (
	"net/http"
	"testing"
)

// Test scanFileHandler, should return 200
func TestScanFileHandler_NonVirus(t *testing.T) {
	want := http.StatusOK
	qParams := make(map[string]string, 1)
	qParams["path"] = "/clamav/tmp/ok/test.txt"
	url, err := getURL(&qParams, "scanFile")
	if err != nil {
		t.Fatalf("TestScanFileHandler_NonVirus failed when creating url, %v", err)
	}
	res, err := c.Get(url.String())
	if err != nil {
		t.Fatalf("TestScanFileHandler_NonVirus failed, wanted %d, got err: %v", want, err)
	}
	got := res.StatusCode
	if got != want {
		t.Fatalf("ScanFileTestScanFileHandler_NonVirus failed, wanted %d, got %d", want, got)
	}
}

// Test scanFileHandler, should return 406
func TestScanFileHandler_WithVirus(t *testing.T) {
	want := http.StatusNotAcceptable
	res, err := c.Get(baseURL.String() + "/scanFile?path=/clamav/tmp/virus/eicar.test")
	if err != nil {
		t.Fatalf("TestScanFileHandler_WithVirus failed, wanted %d, got err: %v", want, err)
	}
	got := res.StatusCode
	if got != want {
		t.Fatalf("TestScanFileHandler_WithVirus failed, wanted %d, got %v", want, got)
	}
}
