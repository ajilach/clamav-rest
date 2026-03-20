package gotest

import (
	"net/http"
	"os"
	"testing"
)

// Test with hello world text file
func TestScanV1_NonVirus(t *testing.T) {
	fName := "/clamav/tmp/ok/test.txt"
	req, err := setupV1(fName)
	if err != nil {
		t.Fatalf("TestScanV1_NonVirus failed when setting up test, %v", err)
	}
	resp, err := c.Do(req)
	if err != nil {
		t.Errorf("TestScanV1_NonVirus failed when calling clamav-rest, %v", err)
	}
	want := http.StatusOK
	got := resp.StatusCode
	if got != want {
		t.Errorf("TestScanV1_NonVirus failed, got %v, but wanted %v", got, want)
	}
}

// Test with eicar.test
func TestScanV1_WithVirus(t *testing.T) {
	fName := "/clamav/tmp/virus/eicar.test"
	req, err := setupV1(fName)
	if err != nil {
		t.Fatalf("TestScanV1_WithVirus failed when setting up test, %v", err)
	}
	resp, err := c.Do(req)
	if err != nil {
		t.Errorf("TestScanV1_WithVirus failed when calling clamav-rest, %v", err)
	}
	want := http.StatusNotAcceptable
	got := resp.StatusCode
	if got != want {
		t.Errorf("TestScanV1_WithVirus failed, got %v, but wanted %v", got, want)
	}
}

// setup
func setupV1(fName string) (*http.Request, error) {
	file, err := os.Open(fName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	req, err := getReqWithFile(file)
	if err != nil {
		return nil, err
	}
	reqURL, err := getURL(nil, "/scan")
	if err != nil {
		return nil, err
	}
	req.URL = reqURL
	return req, nil
}
