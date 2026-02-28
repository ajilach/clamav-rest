package gotest

import (
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
)

func TestFileSizeSlightlyExceeded_RequestEntityTooLarge(t *testing.T) {
	file, err := os.Create("/clamav/tmp/testfile.txt")
	if err != nil {
		t.Fatalf("TestFileSizeSlightlyExceeded_RequestEntityTooLarge failed, unable to create testfile, %v", err)
	}
	defer file.Close()
	_, err = io.CopyN(file, rand.Reader, 10*1024*1024+10) // 10+ MB
	if err != nil {
		t.Fatalf("TestFileSizeSlightlyExceeded_RequestEntityTooLarge failed, unable to write test file, %v", err)
	}
	req, err := getReqWithFile(file)
	if err != nil {
		t.Fatalf("TestFileSizeSlightlyExceeded_RequestEntityTooLarge failed, error creating request from file, %v", err)
	}
	reqURL, err := getURL(nil, "v2", "scan")
	if err != nil {
		t.Fatalf("TestFileSizeSlightlyExceeded_RequestEntityTooLarge failed when creating URL, %v", err)
	}
	req.URL = reqURL
	fmt.Printf("%v\n%v\n%v\n", req.URL.String(), req.Method, req.Header.Get("Content-Type"))
	resp, err := c.Do(req)
	if err != nil {
		t.Errorf("TestFileSizeSlightlyExceeded_RequestEntityTooLarge failed when sending request to Clamav-rest, %v", err)
	}
	want := http.StatusRequestEntityTooLarge
	got := resp.StatusCode
	if got != want {
		t.Errorf("TestFileSizeSlightlyExceeded_RequestEntityTooLarge failed, wanted %d, got %d", want, got)
	}
}
