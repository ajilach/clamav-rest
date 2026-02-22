package gotest

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	baseURL = url.URL{
		Scheme: "http",
		Host:   "localhost:9000",
	}
	c = http.Client{}
)

// Test scanFileHandler, should return 200
func TestScanFileHandler_NonVirus(t *testing.T) {
	res, err := c.Get(baseURL.String() + "/scanFile?path=/clamav/tmp/ok/test.txt")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

// Test scanFileHandler, should return 406
func TestScanFileHAndler_WithVirus(t *testing.T) {
	res, err := c.Get(baseURL.String() + "/scanFile?path=/clamav/tmp/virus/eicar.test")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotAcceptable, res.StatusCode)
}

// Test scanPathHandler, should return 200
func TestScanPathHandler_NonVirus(t *testing.T) {
	res, err := c.Get(baseURL.String() + "/scanPath?path=/clamav/tmp/ok")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

// Test scanPathHandler, should return 406
func TestScanPathHandler_WithVirus(t *testing.T) {
	res, err := c.Get(baseURL.String() + "/scanPath?path=/clamav/tmp/virus")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotAcceptable, res.StatusCode)
}
