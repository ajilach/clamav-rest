package gotest

import (
	"net/http"
	"os"
	"testing"
)

func getResultFromFile(fName string, testName string, t *testing.T) int {
	file, err := os.Open(fName)
	if err != nil {
		t.Fatalf("%v failed when opening test file, %v", testName, err)
	}
	defer file.Close()
	req, err := getReqWithFile(file)
	if err != nil {
		t.Fatalf("%v failed when creating request, %v", testName, err)
	}
	reqURL, err := getURL(nil, "v2", "scan")
	if err != nil {
		t.Fatalf("%v failed when creating request url, %v", testName, err)
	}
	req.URL = reqURL
	resp, err := c.Do(req)
	if err != nil {
		t.Errorf("%v failed when calling clamav-rest, %v", testName, err)
	}
	return resp.StatusCode
}

func TestV2ScanSingleFile_NonVirus(t *testing.T) {
	want := http.StatusOK
	fName := "/clamav/tmp/ok/test.txt"
	got := getResultFromFile(fName, "v2ScanSingleFile_NonVirus", t)
	if got != want {
		t.Errorf("v2ScanSingleFile_NonVirus failed, wanted %v, got %v", want, got)
	}
}

func TestV2ScanSingleFile_WithVirus(t *testing.T) {
	want := http.StatusNotAcceptable
	fName := "/clamav/tmp/virus/eicar.test"
	got := getResultFromFile(fName, "v2ScanSingleFile_WithVirus", t)
	if got != want {
		t.Errorf("v2ScanSingleFile_NonVirus failed, wanted %v, got %v", want, got)
	}
}
