package gotest

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"os"
	"testing"

	"golang.org/x/net/http2"
)

var fileName string = "/clamav/tmp/virus/eicar.test"

// Test HTTP/1.1 over http
func TestHTTP1_1(t *testing.T) {
	ts := http.DefaultTransport.(*http.Transport).Clone()
	ts.Protocols = new(http.Protocols)
	ts.Protocols.SetHTTP1(true)
	ts.Protocols.SetHTTP2(false)

	resp, err := performCall(ts, nil, nil)
	if err != nil {
		t.Errorf("TestHTTP1_1 failed, %v", err)
	}
	gotScheme := resp.Request.URL.Scheme
	wantScheme := "http"
	gotProtoMajor := resp.ProtoMajor
	wantProtoMajor := 1
	gotProtoMinor := resp.ProtoMinor
	wantProtoMinor := 1
	if gotProtoMajor != wantProtoMajor {
		t.Errorf("TestHTTP1_1 failed, wanted major protocol version %v, but got major protocol version %v", wantProtoMajor, gotProtoMajor)
	}
	if gotProtoMinor != wantProtoMinor {
		t.Errorf("TestHTTP1_1 failed, wanted minor protocol version %v, but got minor protocol version %v", wantProtoMinor, gotProtoMinor)
	}
	if gotScheme != wantScheme {
		t.Errorf("TestHTTP1_1 failed, wanted scheme %v, but got scheme %v", wantScheme, gotScheme)
	}
}

// Test h2c, HTTP/2 over HTTP (non-tls) with Prior knowledge
func TestH2C(t *testing.T) {
	// h2c is not implemented in the stdlib http.Transport, if *http.Transport.Protocols.SetHTTP2(true) and SetHTTP1(false) is used, HTTP/2 over TLS will be attempted, or it will fall back to HTTP/1.1 over non-TLS HTTP.
	//ts := &http2.Transport{
	//	AllowHTTP: true,
	//	DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
	//		return net.Dial(network, addr)
	//	},
	//}
	tsp := http.DefaultTransport.(*http.Transport).Clone()
	tsp.Protocols = new(http.Protocols)
	tsp.Protocols.SetUnencryptedHTTP2(true)
	if !tsp.Protocols.UnencryptedHTTP2() {
		t.Errorf("TestH2C failed, tried to set H2C but failed")
	}
	res, err := performCall(tsp, nil, nil)
	if err != nil {
		t.Errorf("TestHTTP2 failed, %v", err)
	}

	gotScheme := res.Request.URL.Scheme
	wantScheme := "http"
	wantProtoMajor := 2
	gotProtoMajor := res.ProtoMajor
	if gotProtoMajor != wantProtoMajor || gotScheme != wantScheme {
		t.Errorf("TestHTTP2 failed, wanted scheme %v, got scheme %v, wanted major protocol version %v, got major protocol version %v", wantScheme, gotScheme, wantProtoMajor, gotProtoMajor)
	}
}

// Test to check that HTTP/2 works over TLS
func TestHTTP2_TLS(t *testing.T) {
	ts := http.DefaultTransport.(*http.Transport).Clone()
	ts.Protocols = new(http.Protocols)
	ts.Protocols.SetHTTP2(true)
	ts.TLSClientConfig = &tls.Config{
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: true,
	}
	url, err := getURL(nil, "v2", "scan")
	if err != nil {
		t.Errorf("TestHTTP2_TLS failed when creating url, %v", err)
	}
	url.Scheme = "https"
	url.Host = net.JoinHostPort(url.Hostname(), "9443")
	res, err := performCall(ts, nil, url)
	if err != nil {
		t.Errorf("TestHTTP2_TLS failed, %v", err)
	}

	gotScheme := res.Request.URL.Scheme
	wantScheme := "https"
	wantProto := 2
	gotProto := res.ProtoMajor
	if gotProto != wantProto || gotScheme != wantScheme {
		t.Errorf("TestHTTP2_TLS failed, wanted scheme %v, got %v, wanted major protocol version %d, got major protocol version %d", wantScheme, gotScheme, wantProto, gotProto)
	}
}

func performCall(ts *http.Transport, ts2 *http2.Transport, url *url.URL) (*http.Response, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	req, err := getReqWithFile(file)
	if err != nil {
		return nil, err
	}
	if url == nil {
		u, err := getURL(nil, "v2", "scan")
		if err != nil {
			return nil, err
		}
		req.URL = u
	} else {
		req.URL = url
	}

	if err != nil {
		return nil, err
	}
	var client *http.Client
	if ts != nil {
		client = &http.Client{Transport: ts}
	} else {
		client = &http.Client{Transport: ts2}
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}
