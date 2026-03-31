package gotest

import (
	"crypto/rand"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"testing"
)

// Test v2/scan with one normal test file and one random file, should return 200 OK
func TestV2ScanMultiFile_NonVirus(t *testing.T) {
	want := http.StatusOK
	fName1 := "/clamav/tmp/ok/test.txt"
	fName2 := "/clamav/tmp/test2.txt"
	reqURL, err := getURL(nil, "v2", "scan")
	if err != nil {
		t.Fatalf("TestV2ScanMultiFile_NonVirus failed when creating URL, %v", err)
	}

	got := multiFileSend(t, reqURL, fName1, fName2)

	if got != want {
		t.Errorf("TestV2ScanMultiFile_NonVirus failed, wanted %v, but got %v", want, got)
	}
}

// Test v2/scan with one clean file and a following virus file, should return 406 Not Acceptable
func TestV2ScanMultiFile_SafeAndVirus(t *testing.T) {
	want := http.StatusNotAcceptable
	fNameOk := "/clamav/tmp/ok/test.txt"
	fNameVirus := "/clamav/tmp/virus/eicar.test"
	reqURL, err := getURL(nil, "v2", "scan")
	if err != nil {
		t.Fatalf("TestV2ScanMultiFile_SafeAndVirus failed when creating URL, %v", err)
	}

	got := multiFileSend(t, reqURL, fNameOk, fNameVirus)

	if got != want {
		t.Errorf("TestV2ScanMultiFile_SafeAndVirus failed, wanted %v, but got %v", want, got)
	}
}

// Test v2/scan with one virus file and a following clean file, should return 406 Not Acceptable
func TestV2ScanMultiFile_VirusAndSafe(t *testing.T) {
	want := http.StatusNotAcceptable
	fNameOk := "/clamav/tmp/ok/test.txt"
	fNameVirus := "/clamav/tmp/virus/eicar.test"
	reqURL, err := getURL(nil, "v2", "scan")
	if err != nil {
		t.Fatalf("TestV2ScanMultiFile_VirusAndSafe failed when creating URL, %v", err)
	}

	got := multiFileSend(t, reqURL, fNameVirus, fNameOk)

	if got != want {
		t.Errorf("TestV2ScanMultiFile_VirusAndSafe failed, wanted %v, but got %v", want, got)
	}
}

// Adds the files to the request in order of slice, sends the request and returns the response code,
func multiFileSend(t *testing.T, reqURL *url.URL, fileNames ...string) int {
	files := []*os.File{}
	for _, f := range fileNames {
		var file *os.File
		_, err := os.Stat(f)
		if errors.Is(err, os.ErrNotExist) {
			file, err = os.Create(f)
			if err != nil {
				t.Fatalf("TestV2ScanMultiFile_NonVirus failed when creating %v, err: %v", f, err)
			}
			_, err = io.CopyN(file, rand.Reader, 10*1024)
			if err != nil {
				t.Fatalf("TestV2ScanMultiFile_NonVirus failed when creating random file %v, err: %v", f, err)
			}
			files = append(files, file)
		} else if err != nil {
			t.Fatalf("TestV2ScanMultiFile_NonVirus failed when creating random file %v, err: %v", f, err)
		} else {
			file, err := os.Open(f)
			files = append(files, file)
			if err != nil {
				t.Fatalf("TestV2ScanMultiFile_NonVirus failed when opening %v, err: %v", f, err)
			}
		}
		defer file.Close()
	}

	req, err := getReqWithFile(files...)
	if err != nil {
		t.Fatalf("TestV2ScanMultiFile_NonVirus failed when creating request, %v", err)
	}

	req.URL = reqURL
	resp, err := c.Do(req)
	if err != nil {
		t.Fatalf("TestV2ScanMultiFile_NonVirus failed when calling clamav-rest, %v", err)
	}
	return resp.StatusCode
}
