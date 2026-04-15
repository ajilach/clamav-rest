package gotest

import (
	"crypto/rand"
	"io"
	"net/http"
	"os"
	"testing"
)

func TestFileSizeLargelyExceeded_RequestEntityTooLarge(t *testing.T) {
	fName := "/clamav/tmp/testfile2.txt"
	file, err := os.Create(fName)
	if err != nil {
		t.Fatalf("TestFileSizeLargelyExceeded_RequestEntityTooLarge failed, unable to create testfile, %v", err)
	}
	defer file.Close()
	defer cleanup("/clamav/tmp/testfile2.txt")
	_, err = io.CopyN(file, rand.Reader, 20*1024*1024)
	if err != nil {
		t.Fatalf("TestFileSizeLargelyExceeded_RequestEntityTooLarge failed, unable to write data to test file, %v", err)
	}
	req, err := getReqWithFile(file)
	if err != nil {
		t.Fatalf("TestFileSizeLargelyExceeded_RequestEntityTooLarge failed, error creating request, %v", err)
	}
	reqURL, err := getURL(nil, "v2", "scan")
	if err != nil {
		t.Fatalf("TestFileSizeLargelyExceeded_RequestEntityTooLarge failed when creating URL, %v", err)
	}
	req.URL = reqURL
	resp, err := c.Do(req)
	if err != nil {
		t.Fatalf("TestFileSizeLargelyExceeded_RequestEntityTooLarge failed when sending request to clamav-rest, %v", err)
	}
	want := http.StatusRequestEntityTooLarge
	got := resp.StatusCode
	if got != want {
		t.Errorf("TestFileSizeLargelyExceeded_RequestEntityTooLarge failed, wanted %d, got %d", want, got)
	}
}
