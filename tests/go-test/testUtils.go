// Package gotest is used to do end-to-end tests written in go, against the clamav-rest api.
// If setup code fails, t.Fatalf() is used, if call to clamav-rest or response fails of is not expected, t.Errorf() is used.
package gotest

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
)

var (
	baseURL = url.URL{
		Scheme: "http",
		Host:   "localhost:9000",
	}
	c = http.Client{}
)

func getReqWithFile(file ...*os.File) (*http.Request, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for _, f := range file {

		// make sure file offset is at the beginning of the file
		_, err := f.Seek(0, io.SeekStart)
		if err != nil {
			return nil, err
		}

		part, err := writer.CreateFormFile("file", f.Name())
		if err != nil {
			return nil, err
		}
		if _, err := io.Copy(part, f); err != nil {
			return nil, err
		}
	}
	err := writer.Close()
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", "", &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, nil
}

func getURL(qParams *map[string]string, pathFragments ...string) (*url.URL, error) {
	addr, err := url.JoinPath(baseURL.String(), pathFragments...)
	if err != nil {
		return nil, err
	}
	reqURL, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}
	if qParams != nil {
		q := reqURL.Query()
		for k, v := range *qParams {
			q.Set(k, v)
		}
		reqURL.RawQuery = q.Encode()
	}
	return reqURL, nil
}

func cleanup(file string) {
	_ = os.Remove(file)
}
