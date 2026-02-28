package gotest

import (
	"net/http"
	"testing"
)

// Test scanPathHandler, should return 200
func TestScanPathHandler_NonVirus(t *testing.T) {
	want := http.StatusOK
	res, err := c.Get(baseURL.String() + "/scanPath?path=/clamav/tmp/ok")
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
	res, err := c.Get(baseURL.String() + "/scanPath?path=/clamav/tmp/virus")
	if err != nil {
		t.Fatalf("TestScanPathHandler_WithVirus failed, wanted %d, got err: %v", want, err)
	}
	got := res.StatusCode
	if got != want {
		t.Fatalf("TestScanPathHandler_NonVirus failed, wanted %d, got %v", want, got)
	}
}
