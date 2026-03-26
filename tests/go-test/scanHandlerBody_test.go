package gotest

import (
	"bufio"
	"net/http"
	"os"
	"testing"
)

// Test scanHandlerBody endpoint with OK file, should return 200
func TestScanHandlerBody_nonVirus(t *testing.T) {
	fName := "/clamav/tmp/ok/test.txt"
	headers := make(map[string]string, 1)
	want := http.StatusOK
	headers["Content-Type"] = "text/plain"
	resp := setupScanHandlerBody(fName, t, headers, want)
	got := resp.StatusCode
	if got != want {
		t.Errorf("TestScanHandlerBody_nonVirus failed, wanted %v, got %v", want, got)
	}
}

// Test scanHandlerBody endpoint with eicar virus test file, should return 406
func TestScanHandlerBody_WithVirus(t *testing.T) {
	fName := "/clamav/tmp/virus/eicar.test"
	headers := make(map[string]string, 1)
	want := http.StatusNotAcceptable
	headers["Content-Type"] = "application/octet-stream"
	resp := setupScanHandlerBody(fName, t, headers, want)
	got := resp.StatusCode
	if got != want {
		t.Errorf("TestScanHandlerBody_nonVirus failed, wanted %v, got %v", want, got)
	}
}

// setup and make call
func setupScanHandlerBody(fName string, t *testing.T, headers map[string]string, want int) *http.Response {
	file, err := os.Open(fName)
	if err != nil {
		t.Fatalf("TestScanHandlerBody_nonVirus failed when opening test file, %v", err)
	}
	defer file.Close()
	reqURL, err := getURL(nil, "scanHandlerBody")
	if err != nil {
		t.Fatalf("TestScanHandlerBody_nonVirus failed, unable to create url, %v", err)
	}
	reader := bufio.NewReader(file)
	req, err := http.NewRequest("POST", reqURL.String(), reader)
	if err != nil {
		t.Fatalf("TestScanHandlerBody_nonVirus failed, unable to create request with fileReader, %v", err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := c.Do(req)
	if err != nil {
		t.Errorf("TestScanHandlerBody_nonVirus failed when sending request to clamav-rest, wanted %v, but got err %v", want, err)
	}
	return resp
}
