package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"path"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/dutchcoders/go-clamd"
)

const (
	defaultScanURLTimeout   = 2 * time.Minute
	maxScanURLRedirects     = 5
	maxScanURLRequestLength = 64 << 10 // 64 KiB is plenty for a JSON body holding a (pre-signed) URL
)

// destinationError is returned when a URL (or one of its redirects) points to a scheme,
// host or IP address that is not allowed by the configuration.
type destinationError struct {
	reason string
}

func (e *destinationError) Error() string {
	return "destination not allowed: " + e.reason
}

func notAllowed(format string, a ...any) error {
	return &destinationError{reason: fmt.Sprintf(format, a...)}
}

// Address ranges that are not covered by the netip.Addr helper methods but must not be
// reachable unless private networks are explicitly allowed.
var extraBlockedPrefixes = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),     // "this" network
	netip.MustParsePrefix("100.64.0.0/10"), // carrier-grade NAT, also used for some cloud metadata services
	netip.MustParsePrefix("192.0.0.0/24"),  // IETF protocol assignments
	netip.MustParsePrefix("198.18.0.0/15"), // benchmarking
	netip.MustParsePrefix("240.0.0.0/4"),   // reserved, includes broadcast
}

type scanURLConfig struct {
	// allowPrivateNetworks allows fetching from loopback, private, link-local and other
	// non-public addresses, e.g. a MinIO instance inside the same cluster.
	allowPrivateNetworks bool
	// allowedHosts restricts the hosts that may be fetched from. Entries are either an exact
	// host name or a wildcard of the form "*.example.com". Empty means any host.
	allowedHosts []string
	// timeout limits the total time for downloading and scanning a single URL.
	timeout time.Duration
	// forbid scanURL endpoint
	allowScanURL bool 
}

type scanURLRequest struct {
	URL string `json:"url"`
}

type urlScanner struct {
	cfg       scanURLConfig
	client    *http.Client
	clamdAddr string
}

func loadScanURLConfig(opts map[string]string) scanURLConfig {
	cfg := scanURLConfig{timeout: defaultScanURLTimeout}

	if v := opts["SCAN_URL_ALLOW_PRIVATE_NETWORKS"]; v != "" {
		allow, err := strconv.ParseBool(v)
		if err != nil {
			log.Printf("Invalid value for SCAN_URL_ALLOW_PRIVATE_NETWORKS %q, private networks stay blocked\n", v)
		}
		cfg.allowPrivateNetworks = allow
	}

	for _, h := range strings.Split(opts["SCAN_URL_ALLOWED_HOSTS"], ";") {
		h = strings.ToLower(strings.TrimSpace(h))
		if h != "" {
			cfg.allowedHosts = append(cfg.allowedHosts, h)
		}
	}

	if v := opts["SCAN_URL_TIMEOUT"]; v != "" {
		timeout, err := time.ParseDuration(v)
		if err != nil || timeout <= 0 {
			log.Printf("Invalid value for SCAN_URL_TIMEOUT %q, using default of %v\n", v, defaultScanURLTimeout)
		} else {
			cfg.timeout = timeout
		}
	}
	if v := opts["SCAN_URL_ALLOW"]; v != "" {
		allow, err := strconv.ParseBool(v)
		if err != nil {
			log.Printf("Invalid value for SCAN_URL_ALLOW, defaulting to allow\n")
			allow = true
		}
		cfg.allowScanURL = allow
	}

	return cfg
}

func newURLScanner(cfg scanURLConfig, clamdAddr string) *urlScanner {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		// Control is called after DNS resolution for every connection attempt, which also
		// covers redirects and protects against DNS rebinding.
		Control: func(network, address string, _ syscall.RawConn) error {
			if cfg.allowPrivateNetworks {
				return nil
			}
			host, _, err := net.SplitHostPort(address)
			if err != nil {
				return err
			}
			ip, err := netip.ParseAddr(host)
			if err != nil {
				return err
			}
			if isNonPublicAddr(ip) {
				return notAllowed("%s is not a public address", ip)
			}
			return nil
		},
	}

	transport := &http.Transport{
		// Proxies are deliberately not used, otherwise the address checks above would only
		// apply to the proxy and not to the actual destination.
		Proxy:                 nil,
		DialContext:           dialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   cfg.timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= maxScanURLRedirects {
				return fmt.Errorf("stopped after %d redirects", maxScanURLRedirects)
			}
			return cfg.validateURL(req.URL)
		},
	}

	return &urlScanner{cfg: cfg, client: client, clamdAddr: clamdAddr}
}

// validateURL checks the scheme and, if configured, the host allow list.
func (cfg scanURLConfig) validateURL(u *url.URL) error {
	if u.Scheme != "http" && u.Scheme != "https" {
		return notAllowed("unsupported scheme %q, only http and https are allowed", u.Scheme)
	}
	if u.Hostname() == "" {
		return notAllowed("url has no host")
	}
	if !cfg.hostAllowed(u.Hostname()) {
		return notAllowed("host %q is not in SCAN_URL_ALLOWED_HOSTS", u.Hostname())
	}
	return nil
}

func (cfg scanURLConfig) hostAllowed(host string) bool {
	if len(cfg.allowedHosts) == 0 {
		return true
	}
	host = strings.ToLower(strings.TrimSuffix(host, "."))
	for _, allowed := range cfg.allowedHosts {
		if suffix, ok := strings.CutPrefix(allowed, "*."); ok {
			if strings.HasSuffix(host, "."+suffix) {
				return true
			}
		} else if host == allowed {
			return true
		}
	}
	return false
}

func isNonPublicAddr(ip netip.Addr) bool {
	ip = ip.Unmap()
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() ||
		ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() ||
		ip.IsInterfaceLocalMulticast() || ip.IsMulticast() {
		return true
	}
	for _, p := range extraBlockedPrefixes {
		if p.Contains(ip) {
			return true
		}
	}
	return false
}

// redactURL removes query string, fragment and user info, so that credentials contained in
// e.g. pre-signed URLs never end up in logs or responses.
func redactURL(u *url.URL) string {
	r := url.URL{Scheme: u.Scheme, Host: u.Host, Path: u.Path}
	return r.String()
}

// fileNameFromURL returns the last path segment of the URL, or the host if there is none.
func fileNameFromURL(u *url.URL) string {
	name := path.Base(u.Path)
	if name == "/" || name == "." || name == "" {
		return u.Hostname()
	}
	return name
}

// errRecordingReader remembers the first read error that is not io.EOF. clamd's ScanStream
// stops reading on any error and still scans what it got so far, so without this a
// failed download would be reported as a clean file.
type errRecordingReader struct {
	r   io.Reader
	err error
}

func (e *errRecordingReader) Read(p []byte) (int, error) {
	n, err := e.r.Read(p)
	if err != nil && err != io.EOF && e.err == nil {
		e.err = err
	}
	return n, err
}

func writeScanResponse(w http.ResponseWriter, resp scanResponse) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(resp.httpStatus)
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Error marshalling json, %v\n", err)
		return
	}
	if _, err := w.Write(jsonResp); err != nil {
		log.Printf("unable to write response to client, %v\n", err)
	}
}

func errorResponse(httpStatus int, fileName, description string) scanResponse {
	return scanResponse{Status: clamd.RES_ERROR, Description: description, FileName: fileName, httpStatus: httpStatus}
}

// fetchErrorResponse maps an error that occurred while downloading to a response.
func fetchErrorResponse(err error, fileName string) scanResponse {
	var destErr *destinationError
	var netErr net.Error
	switch {
	case errors.As(err, &destErr):
		return errorResponse(http.StatusForbidden, fileName, "url not allowed, "+destErr.reason)
	case errors.Is(err, context.DeadlineExceeded), errors.As(err, &netErr) && netErr.Timeout():
		return errorResponse(http.StatusGatewayTimeout, fileName, "timeout while downloading the file")
	default:
		return errorResponse(http.StatusBadGateway, fileName, "unable to download the file")
	}
}

// handle downloads the file behind the URL in the request body and streams it to clamd.
func (s *urlScanner) handle(w http.ResponseWriter, r *http.Request) {
	if !s.cfg.allowScanURL {
		writeScanResponse(w,, errorResponse(http.StatusForbidden, "", "forbidden"))
		return
	}
	var req scanURLRequest
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxScanURLRequestLength))
	if err := dec.Decode(&req); err != nil {
		log.Printf("scanUrl was called with an invalid body, %v\n", err)
		writeScanResponse(w, errorResponse(http.StatusBadRequest, "", `request body must be a JSON object like {"url": "https://..."}`))
		return
	}
	if strings.TrimSpace(req.URL) == "" {
		log.Println("scanUrl was called without a url")
		writeScanResponse(w, errorResponse(http.StatusBadRequest, "", "property 'url' is missing"))
		return
	}

	target, err := url.Parse(strings.TrimSpace(req.URL))
	if err != nil {
		writeScanResponse(w, errorResponse(http.StatusBadRequest, "", "property 'url' is not a valid url"))
		return
	}
	fileName := fileNameFromURL(target)
	if err := s.cfg.validateURL(target); err != nil {
		log.Printf("Refusing to scan %v, %v\n", redactURL(target), err)
		writeScanResponse(w, fetchErrorResponse(err, fileName))
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), s.cfg.timeout)
	defer cancel()
	dlReq, err := http.NewRequestWithContext(ctx, http.MethodGet, target.String(), nil)
	if err != nil {
		writeScanResponse(w, errorResponse(http.StatusBadRequest, fileName, "property 'url' is not a valid url"))
		return
	}

	log.Printf("Started downloading %v\n", redactURL(target))
	dlResp, err := s.client.Do(dlReq)
	if err != nil {
		log.Printf("Unable to download %v, %v\n", redactURL(target), errors.Unwrap(err))
		writeScanResponse(w, fetchErrorResponse(err, fileName))
		return
	}
	defer dlResp.Body.Close()

	if dlResp.StatusCode < 200 || dlResp.StatusCode > 299 {
		log.Printf("Unable to download %v, remote server responded with %v\n", redactURL(target), dlResp.Status)
		writeScanResponse(w, errorResponse(http.StatusBadGateway, fileName, fmt.Sprintf("remote server responded with %v", dlResp.Status)))
		return
	}

	c := clamd.NewClamd(s.clamdAddr)
	body := &errRecordingReader{r: dlResp.Body}
	abort := make(chan bool)
	// closing abort lets go-clamd close the connection and end its watcher goroutine
	defer close(abort)

	log.Printf("Started scanning %v\n", redactURL(target))
	results, err := c.ScanStream(body, abort)
	if err != nil {
		if body.err != nil {
			log.Printf("Unable to download %v, %v\n", redactURL(target), body.err)
			writeScanResponse(w, fetchErrorResponse(body.err, fileName))
			return
		}
		log.Printf("Clamd returned an error, probably a too large file as input (causing broken pipe and closed connection) %v\n", err)
		writeScanResponse(w, scanResponse{Status: clamd.RES_PARSE_ERROR, Description: "File size limit exceeded", FileName: fileName, httpStatus: http.StatusRequestEntityTooLarge})
		return
	}

	var resp *scanResponse
	for s := range results {
		if resp != nil {
			continue // drain the channel, only the first result is relevant
		}
		resp = &scanResponse{Status: s.Status, Description: s.Description, FileName: fileName, httpStatus: getHTTPStatusByClamStatus(s)}
		log.Printf("Scan result for %v: %v\n", redactURL(target), s)
	}

	switch {
	case resp != nil && resp.Status == clamd.RES_FOUND:
		// a virus found in a partial download is still a virus
		writeScanResponse(w, *resp)
	case body.err != nil:
		log.Printf("Download of %v was interrupted, %v\n", redactURL(target), body.err)
		writeScanResponse(w, fetchErrorResponse(body.err, fileName))
	case resp == nil:
		writeScanResponse(w, errorResponse(http.StatusInternalServerError, fileName, "no response from clamd"))
	default:
		writeScanResponse(w, *resp)
	}
	log.Printf("Finished scanning %v\n", redactURL(target))
}
