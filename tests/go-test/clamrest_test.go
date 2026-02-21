package gotest

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	baseUrl = url.URL{
		Scheme: "http",
		Host:   "localhost:9000",
	}
	c = http.Client{}
)

func TestScanFileHandler_nonVirus(t *testing.T) {
	res, err := c.Get(baseUrl.String() + "/scanFile?path=/clamav/tmp/ok/test.txt")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Error(t, err)
}

func TestScanFileHAndler_WithVirus(t *testing.T) {
	res, err := c.Get(baseUrl.String() + "/scanFile?path=/clamav/tmp/virus/eicar.test")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotAcceptable, res.StatusCode)
	assert.Error(t, err)
}
