package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"net/url"
	"strings"
	"testing"
	"time"
)

// startFakeClamd starts a minimal clamd that understands INSTREAM and reports a virus if
// the streamed content contains "EICAR". It returns the address in go-clamd notation.
func startFakeClamd(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("unable to start fake clamd, %v", err)
	}
	t.Cleanup(func() { ln.Close() })

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(conn net.Conn) {
				defer conn.Close()
				r := bufio.NewReader(conn)
				cmd, err := r.ReadString('\n')
				if err != nil || cmd != "nINSTREAM\n" {
					return
				}
				var content bytes.Buffer
				for {
					var size uint32
					if err := binary.Read(r, binary.BigEndian, &size); err != nil {
						return
					}
					if size == 0 {
						break
					}
					if _, err := io.CopyN(&content, r, int64(size)); err != nil {
						return
					}
				}
				if strings.Contains(content.String(), "EICAR") {
					io.WriteString(conn, "stream: Eicar-Test-Signature FOUND\n")
				} else {
					io.WriteString(conn, "stream: OK\n")
				}
			}(conn)
		}
	}()

	return "tcp://" + ln.Addr().String()
}

func doScanURL(t *testing.T, s *urlScanner, body string) (*httptest.ResponseRecorder, scanResponse) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/scanUrl", strings.NewReader(body))
	rec := httptest.NewRecorder()
	s.handle(rec, req)
	var resp scanResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response is not valid json, %v: %s", err, rec.Body.String())
	}
	return rec, resp
}

func urlBody(u string) string {
	b, _ := json.Marshal(scanURLRequest{URL: u})
	return string(b)
}

func TestIsNonPublicAddr(t *testing.T) {
	tests := map[string]bool{
		"127.0.0.1":        true,
		"10.1.2.3":         true,
		"172.16.0.1":       true,
		"192.168.1.1":      true,
		"169.254.169.254":  true,
		"100.100.100.200":  true,
		"0.0.0.0":          true,
		"255.255.255.255":  true,
		"::1":              true,
		"fd00:ec2::254":    true,
		"fe80::1":          true,
		"::ffff:127.0.0.1": true,
		"::ffff:10.0.0.1":  true,
		"52.219.170.1":     false,
		"8.8.8.8":          false,
		"2600:1f18::1":     false,
	}
	for ip, want := range tests {
		if got := isNonPublicAddr(netip.MustParseAddr(ip)); got != want {
			t.Errorf("isNonPublicAddr(%s) = %v, want %v", ip, got, want)
		}
	}
}

func TestHostAllowed(t *testing.T) {
	cfg := scanURLConfig{allowedHosts: []string{"minio.internal", "*.amazonaws.com"}, allowScanURL: true}
	tests := map[string]bool{
		"minio.internal": true,
		"MINIO.internal": true,
		"my-bucket.s3.eu-central-1.amazonaws.com": true,
		"amazonaws.com":      false,
		"evilamazonaws.com":  false,
		"sub.minio.internal": false,
		"example.com":        false,
	}
	for host, want := range tests {
		if got := cfg.hostAllowed(host); got != want {
			t.Errorf("hostAllowed(%s) = %v, want %v", host, got, want)
		}
	}

	if !(scanURLConfig{}).hostAllowed("anything.example.com") {
		t.Error("an empty allow list should allow every host")
	}
}

func TestLoadScanURLConfig(t *testing.T) {
	cfg := loadScanURLConfig(map[string]string{})
	if cfg.allowPrivateNetworks || len(cfg.allowedHosts) != 0 || cfg.timeout != defaultScanURLTimeout || !cfg.allowScanURL {
		t.Errorf("unexpected defaults %+v", cfg)
	}

	cfg = loadScanURLConfig(map[string]string{
		"SCAN_URL_ALLOW_PRIVATE_NETWORKS": "true",
		"SCAN_URL_ALLOWED_HOSTS":          " minio.internal ;*.Amazonaws.com;",
		"SCAN_URL_TIMEOUT":                "30s",
		"SCAN_URL_ALLOW":                  "false",
	})
	if !cfg.allowPrivateNetworks {
		t.Error("expected private networks to be allowed")
	}
	if cfg.allowScanURL {
		t.Error("expected allowScanURL to be false")
	}
	if len(cfg.allowedHosts) != 2 || cfg.allowedHosts[0] != "minio.internal" || cfg.allowedHosts[1] != "*.amazonaws.com" {
		t.Errorf("unexpected allowed hosts %q", cfg.allowedHosts)
	}
	if cfg.timeout != 30*time.Second {
		t.Errorf("unexpected timeout %v", cfg.timeout)
	}

	cfg = loadScanURLConfig(map[string]string{"SCAN_URL_ALLOW_PRIVATE_NETWORKS": "yes please", "SCAN_URL_TIMEOUT": "-1s"})
	if cfg.allowPrivateNetworks || cfg.timeout != defaultScanURLTimeout {
		t.Errorf("invalid values should fall back to safe defaults, got %+v", cfg)
	}
}

func TestFileNameAndRedaction(t *testing.T) {
	u, _ := url.Parse("https://user:pass@bucket.s3.amazonaws.com/uploads/My%20File.pdf?X-Amz-Signature=secret#frag")
	if got := fileNameFromURL(u); got != "My File.pdf" {
		t.Errorf("fileNameFromURL = %q", got)
	}
	if got := redactURL(u); strings.Contains(got, "secret") || strings.Contains(got, "pass") || strings.Contains(got, "frag") {
		t.Errorf("redactURL leaks credentials: %q", got)
	}
	u, _ = url.Parse("https://example.com")
	if got := fileNameFromURL(u); got != "example.com" {
		t.Errorf("fileNameFromURL without path = %q", got)
	}
}

func TestScanURL_Results(t *testing.T) {
	clamdAddr := startFakeClamd(t)
	files := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ok.txt":
			io.WriteString(w, "hello world")
		case "/eicar.test":
			io.WriteString(w, "X5O!P%@AP[4\\PZX54(P^)7CC)7}$EICAR-STANDARD-ANTIVIRUS-TEST-FILE!$H+H*")
		case "/redirect":
			http.Redirect(w, r, "/ok.txt", http.StatusFound)
		case "/truncated":
			// announce more content than is sent, so the download fails midway
			w.Header().Set("Content-Length", "1000")
			io.WriteString(w, "partial")
		case "/slow":
			time.Sleep(500 * time.Millisecond)
			io.WriteString(w, "too late")
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(files.Close)

	s := newURLScanner(scanURLConfig{allowPrivateNetworks: true, timeout: 200 * time.Millisecond, allowScanURL: true}, clamdAddr)

	tests := []struct {
		path       string
		wantHTTP   int
		wantStatus string
		wantName   string
	}{
		{"/ok.txt?X-Amz-Signature=secret", http.StatusOK, "OK", "ok.txt"},
		{"/eicar.test", http.StatusNotAcceptable, "FOUND", "eicar.test"},
		{"/redirect", http.StatusOK, "OK", "redirect"},
		{"/missing.txt", http.StatusBadGateway, "ERROR", "missing.txt"},
		{"/truncated", http.StatusBadGateway, "ERROR", "truncated"},
		{"/slow", http.StatusGatewayTimeout, "ERROR", "slow"},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			rec, resp := doScanURL(t, s, urlBody(files.URL+tt.path))
			if rec.Code != tt.wantHTTP || resp.Status != tt.wantStatus || resp.FileName != tt.wantName {
				t.Errorf("got %d %+v, want %d %s %s", rec.Code, resp, tt.wantHTTP, tt.wantStatus, tt.wantName)
			}
			if strings.Contains(rec.Body.String(), "secret") {
				t.Errorf("response leaks the query string: %s", rec.Body.String())
			}
		})
	}
}

func TestScanURL_Rejected(t *testing.T) {
	files := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirect-elsewhere" {
			http.Redirect(w, r, "http://localhost:1/ok.txt", http.StatusFound)
			return
		}
		io.WriteString(w, "hello world")
	}))
	t.Cleanup(files.Close)

	// clamd must never be reached in these tests
	const noClamd = "tcp://127.0.0.1:1"
	defaults := newURLScanner(scanURLConfig{timeout: time.Second, allowScanURL: true}, noClamd)
	allowList := newURLScanner(scanURLConfig{allowPrivateNetworks: true, allowedHosts: []string{"127.0.0.1"}, timeout: time.Second, allowScanURL: true}, noClamd)

	tests := []struct {
		name     string
		scanner  *urlScanner
		body     string
		wantHTTP int
	}{
		{"invalid json", defaults, "not json", http.StatusBadRequest},
		{"missing url", defaults, `{}`, http.StatusBadRequest},
		{"unsupported scheme", defaults, urlBody("file:///etc/passwd"), http.StatusForbidden},
		{"no host", defaults, urlBody("http:///foo"), http.StatusForbidden},
		{"loopback blocked by default", defaults, urlBody(files.URL + "/ok.txt?X-Amz-Signature=secret"), http.StatusForbidden},
		{"host not in allow list", allowList, urlBody(strings.Replace(files.URL, "127.0.0.1", "localhost", 1)), http.StatusForbidden},
		{"redirect to host not in allow list", allowList, urlBody(files.URL + "/redirect-elsewhere"), http.StatusForbidden},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec, resp := doScanURL(t, tt.scanner, tt.body)
			if rec.Code != tt.wantHTTP || resp.Status != "ERROR" {
				t.Errorf("got %d %+v, want %d ERROR", rec.Code, resp, tt.wantHTTP)
			}
			if strings.Contains(rec.Body.String(), "secret") {
				t.Errorf("response leaks the query string: %s", rec.Body.String())
			}
		})
	}
}
