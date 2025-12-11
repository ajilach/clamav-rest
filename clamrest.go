package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dutchcoders/go-clamd"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
)

var opts map[string]string

var noOfFoundViruses = prometheus.NewCounter(prometheus.CounterOpts{
	Name: "no_of_found_viruses",
	Help: "The total number of found viruses",
})

var noOfHitsOnDeprecatedScanEndpoint = prometheus.NewCounter(prometheus.CounterOpts{
	Name: "no_of_hits_on_deprecated_scan_endpoint",
	Help: "The number of hits on the deprecated /scan endpoint. If this is not 0, inform your clients to ajust their code to use the /v2/scan endpoint instead.",
})

func init() {
	log.SetOutput(io.Discard)
}

func clamversion(w http.ResponseWriter, r *http.Request) {
	c := clamd.NewClamd(opts["CLAMD_PORT"])

	version, err := c.Version()
	if err != nil {
		errJSON, eErr := json.Marshal(err)
		if eErr != nil {
			log.Println(eErr)
			return
		}
		fmt.Fprint(w, string(errJSON))
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	for versionStr := range version {
		if strings.HasPrefix(versionStr.Raw, "ClamAV ") {
			versionValues := strings.Split(strings.Replace(versionStr.Raw, "ClamAV ", "", 1), "/")
			respJSON := fmt.Sprintf("{ \"Clamav\": \"%s\" }", versionValues[0])
			if len(versionValues) == 3 {
				respJSON = fmt.Sprintf("{ \"Clamav\": \"%s\", \"Signature\": \"%s\" , \"Signature_date\": \"%s\" }", versionValues[0], versionValues[1], versionValues[2])
			}
			fmt.Fprint(w, string(respJSON))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	c := clamd.NewClamd(opts["CLAMD_PORT"])

	response, err := c.Stats()
	if err != nil {
		errJSON, eErr := json.Marshal(err)
		if eErr != nil {
			log.Println(eErr)
			return
		}
		fmt.Fprint(w, string(errJSON))
		return
	}

	resJSON, eRes := json.Marshal(response)
	if eRes != nil {
		log.Println(eRes)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprint(w, string(resJSON))
}

func scanFileHandler(w http.ResponseWriter, r *http.Request) {
	paths, ok := r.URL.Query()["path"]
	if !ok || len(paths[0]) < 1 {
		log.Println("scanFile was called without a path")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("URL param 'path' is missing"))
		return
	}

	path := paths[0]

	c := clamd.NewClamd(opts["CLAMD_PORT"])
	log.Printf("Started scanning %v\n", path)
	response, err := c.ScanFile(path)
	if err != nil {
		errJSON, marshalErr := json.Marshal(err)
		if marshalErr != nil {
			log.Printf("error marshalling error from clamd, %v\n", marshalErr)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("error from clamd when scanning file"))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, string(errJSON))
		return
	}
	var resp []scanResponse
	// loop over the channel to collect the response
	for respItem := range response {
		scanResp := scanResponse{
			httpStatus:  getHTTPStatusByClamStatus(respItem),
			Status:      respItem.Status,
			Description: respItem.Description,
			FileName:    path,
		}
		if respItem.Status == clamd.RES_PARSE_ERROR {
			scanResp.Description += ", this likely means the file path supplied to the api doesn't point to a file on disk."
		}
		resp = append(resp, scanResp)
	}

	// if the file is not found, we will get two almost identical error scanResponses on the channel.
	// Will just use the first and call it a day.
	respJSON, err := json.Marshal(resp[0])
	if err != nil {
		log.Printf("error marshalling response to json, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("unable to marshal response"))
		return
	}

	log.Printf("Finished scanning %v\n", path)
	w.WriteHeader(resp[0].httpStatus)
	w.Write(respJSON)
}

func scanPathHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("url query: " + r.URL.RawQuery)
	paths, ok := r.URL.Query()["path"]
	if !ok || len(paths[0]) < 1 {
		log.Println("Url Param 'path' is missing")
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte("URL param 'path' is missing"))
		if err != nil {
			log.Printf("unable to write error msg to client, %v\n", err)
		}
		return
	}

	path := paths[0]

	c := clamd.NewClamd(opts["CLAMD_PORT"])
	response, err := c.AllMatchScanFile(path)
	if err != nil {
		errJSON, eErr := json.Marshal(err)
		if eErr != nil {
			log.Println(eErr)
			return
		}
		fmt.Fprint(w, string(errJSON))
		return
	}

	scanResults := []scanResponse{}
	for responseItem := range response {
		eachResp := scanResponse{Status: responseItem.Status, Description: responseItem.Description}
		eachResp.httpStatus = getHTTPStatusByClamStatus(responseItem)
		scanResults = append(scanResults, eachResp)
	}

	resJSON, eRes := json.Marshal(scanResults)
	if eRes != nil {
		log.Println(eRes)
		return
	}
	w.WriteHeader(getResponseStatus(scanResults))
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprint(w, string(resJSON))
}

func v2ScanHandler(w http.ResponseWriter, r *http.Request) {
	scanner(w, r, 2)
}

// old endpoint version, set deprecation header to indicate usage of the new /v2/scan
func scanHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Deprecation", "version=v1")
	v2url := fmt.Sprintf("%s%s/v2/scan", string(r.URL.Scheme), r.Host)
	w.Header().Add("Link", fmt.Sprintf("%v; rel=successor-version", v2url))
	noOfHitsOnDeprecatedScanEndpoint.Inc()

	scanner(w, r, 1)
}

// This is where the action happens.
func scanner(w http.ResponseWriter, r *http.Request, version int) {
	switch r.Method {
	// POST takes the uploaded file(s) and saves it to disk.
	case "POST":
		c := clamd.NewClamd(opts["CLAMD_PORT"])
		// get the multipart reader for the request.
		reader, err := r.MultipartReader()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// copy each part to destination.
		resp := []scanResponse{}
		for {
			part, err := reader.NextPart()
			if err == io.EOF {
				break
			}

			// if part.FileName() is empty, skip this iteration.
			if part.FileName() == "" {
				if version == 2 {
					fileResp := scanResponse{Status: "ERROR", Description: "MimePart FileName missing", httpStatus: 422}
					resp = append(resp, fileResp)
					log.Println("Not scanning, MimePart FileName not supplied")
				}
				continue
			}

			log.Printf("Started scanning: %v\n", part.FileName())
			var abort chan bool
			response, err := c.ScanStream(part, abort)
			if err != nil {
				// error occurred, response is nil, create a custom response and send it on the channel to handle it together with the other errors.
				response = make(chan *clamd.ScanResult)
				scanErrResult := &clamd.ScanResult{Status: clamd.RES_PARSE_ERROR, Description: "File size limit exceeded"}
				go func() {
					response <- scanErrResult
					close(response)
					log.Printf("Clamd returned an error, probably a too large file as input (causing broken pipe and closed connection) %v\n", err)
					// The underlying service closes the connection if the file is to large, logging output
					// We never receive the clamd output of `^INSTREAM: Size limit reached` up here, just a closed connection.
				}()

			}
			for s := range response {
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				eachResp := scanResponse{Status: s.Status, Description: s.Description}
				if version == 2 {
					eachResp.FileName = part.FileName()
					log.Printf("Scanned file %v\n", part.FileName())
				}
				// Set each possible status and then send the most appropriate one
				eachResp.httpStatus = getHTTPStatusByClamStatus(s)
				resp = append(resp, eachResp)
				log.Printf("Scan result for: %v, %v\n", part.FileName(), s)
			}
			log.Printf("Finished scanning: %v\n", part.FileName())
		}
		w.WriteHeader(getResponseStatus(resp))
		if version == 2 {
			jsonRes, jErr := json.Marshal(resp)
			if jErr != nil {
				log.Printf("Error marshalling json, %v\n", jErr)
			}
			fmt.Fprint(w, string(jsonRes))
		} else {
			for _, v := range resp {
				jsonRes, jErr := json.Marshal(v)
				if jErr != nil {
					log.Printf("Error marshalling json, %v\n", jErr)
				}
				fmt.Fprint(w, string(jsonRes))
			}
		}

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func getHTTPStatusByClamStatus(result *clamd.ScanResult) int {
	switch result.Status {
	case clamd.RES_OK:
		return http.StatusOK // 200
	case clamd.RES_FOUND:
		log.Println("Virus FOUND")
		noOfFoundViruses.Inc()
		return http.StatusNotAcceptable // 406
	case clamd.RES_ERROR:
		return http.StatusBadRequest // 400
	case clamd.RES_PARSE_ERROR:
		if result.Description == "File size limit exceeded" {
			return http.StatusRequestEntityTooLarge // 413
		} else {
			return http.StatusPreconditionFailed // 412
		}
	default:
		return http.StatusNotImplemented // 501
	}
}

// this func returns 406 if one file contains a virus
func getResponseStatus(responses []scanResponse) int {
	result := 200
	for _, r := range responses {
		switch r.httpStatus {
		case 406:
			// uptick the prometheus counter for detected viruses.
			noOfFoundViruses.Inc()
			// early return if virus is found
			return 406
		case 400:
			result = 400
		case 412:
			result = 412
		case 413:
			result = 413
		case 422:
			result = 422
		case 501:
			result = 501
		}
	}

	return result
}

func scanHandlerBody(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	c := clamd.NewClamd(opts["CLAMD_PORT"])

	log.Println("Started scanning plain body")
	var abort chan bool
	defer r.Body.Close()
	response, err := c.ScanStream(r.Body, abort)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err != nil {
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		resp := scanResponse{Status: clamd.RES_PARSE_ERROR, Description: "File size limit exceeded"}
		log.Printf("Clamd returned error, broken pipe and closed connection can indicate too large file, %v\n", err)
		jsonResp, jErr := json.Marshal(resp)
		if jErr != nil {
			log.Printf("Error marshalling json, %v\n", jErr)
		}
		fmt.Fprint(w, string(jsonResp))
		return
	}
	for s := range response {

		resp := scanResponse{Status: s.Status, Description: s.Description}
		// respJson := fmt.Sprintf("{ Status: %q, Description: %q }", s.Status, s.Description)
		resp.httpStatus = getHTTPStatusByClamStatus(s)

		resps := []scanResponse{}
		resps = append(resps, resp)
		w.WriteHeader(getResponseStatus(resps))
		fmt.Fprint(w, resp)
		log.Printf("Scan result for plain body: %v\n", s)
	}
}

func waitForClamD(port string, times int, maxTimes int) {
	clamdTest := clamd.NewClamd(port)
	err := clamdTest.Ping()
	if err != nil {
		log.Println("Clamd did not respond to ping")
	}
	version, err := clamdTest.Version()

	if err != nil {
		if times < maxTimes {
			log.Printf("clamD not running, waiting times [%v]\n", times)
			time.Sleep(time.Second * 4)
			waitForClamD(port, times+1, maxTimes)
		} else {
			log.Printf("Error getting clamd version: %v\n", err)
			os.Exit(1)
		}
	} else {
		for versionString := range version {
			log.Printf("Clamd version: %#v\n", versionString.Raw)
		}
	}
}

func main() {
	opts = make(map[string]string)

	log.SetFlags(0)
	log.SetOutput(new(logWriter))
	// https://github.com/prometheus/client_golang/blob/main/examples/gocollector/main.go
	reg := prometheus.NewRegistry()
	reg.MustRegister(collectors.NewBuildInfoCollector())
	reg.MustRegister(collectors.NewGoCollector(
		collectors.WithGoCollections(collectors.GoRuntimeMemStatsCollection | collectors.GoRuntimeMetricsCollection),
	))
	reg.MustRegister(noOfFoundViruses)
	reg.MustRegister(noOfHitsOnDeprecatedScanEndpoint)

	for _, e := range os.Environ() {
		pair := strings.Split(e, "=")
		opts[pair[0]] = pair[1]
	}

	if opts["CLAMD_PORT"] == "" {
		opts["CLAMD_PORT"] = "tcp://localhost:3310"
	}

	log.Println("Starting clamav rest bridge")
	log.Printf("Connecting to clamd on %v\n", opts["CLAMD_PORT"])

	maxReconnect, err := strconv.Atoi(opts["MAX_RECONNECT_TIME"])
	if err != nil {
		log.Printf("Error converting MAX_RECONNECT_TIME to integer: %v\n", err)
	}

	waitForClamD(opts["CLAMD_PORT"], 1, maxReconnect)

	log.Printf("Connected to clamd on %v\n", opts["CLAMD_PORT"])
	mux := http.NewServeMux()
	// Add cors middleware
	c := cors.New(getCorsPolicy())

	mux.HandleFunc("POST /scan", scanHandler)
	mux.HandleFunc("POST /v2/scan", v2ScanHandler)
	mux.HandleFunc("GET /scanFile", scanFileHandler)
	mux.HandleFunc("GET /scanPath", scanPathHandler)
	mux.HandleFunc("POST /scanHandlerBody", scanHandlerBody)
	mux.HandleFunc("GET /version", clamversion)
	mux.HandleFunc("GET /", home)

	// Prometheus metrics
	mux.Handle("GET /metrics", promhttp.HandlerFor(
		reg,
		promhttp.HandlerOpts{
			// Opt into OpenMetrics to support exemplars.
			EnableOpenMetrics: true,
		},
	))

	// Attach the cors middleware to the middleware chain/request pipeline
	handler := c.Handler(mux)

	// Configure the HTTPS server
	tlsServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", opts["SSL_PORT"]),
		Handler: handler,
	}

	// Configure the HTTP server with h2c support
	var protocols http.Protocols
	protocols.SetHTTP1(true)
	protocols.SetUnencryptedHTTP2(true) // Enable h2c support
	protocols.SetHTTP2(true)

	httpServer := &http.Server{
		Addr:      fmt.Sprintf(":%s", opts["PORT"]),
		Handler:   handler,
		Protocols: &protocols,
	}
	// Start the HTTPS server in a goroutine
	go func() {
		log.Fatal(tlsServer.ListenAndServeTLS("/etc/ssl/clamav-rest/server.crt", "/etc/ssl/clamav-rest/server.key"))
	}()
	// Start the HTTP server
	log.Fatal(httpServer.ListenAndServe())
}

func getCorsPolicy() cors.Options {
	envs := os.Environ()
	// Ignoring Go's naming conventions of non-snake_case naming to keep the same variable name as the env var.
	var allow_origins []string

	// Only allow same-origin requests by default
	// This effectively disables CORS when no origins are explicitly allowed
	for _, env := range envs {
		e := strings.Split(env, "=")
		if strings.EqualFold(e[0], "allow_origins") {
			allow_origins = strings.Split(e[1], ";")
		}
	}

	return cors.Options{
		AllowedOrigins:   allow_origins,
		AllowedHeaders:   []string{"Content-Type", "Accept-Language", "Accept"},
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodOptions},
		AllowCredentials: false,
	}
}

// format logger so it logs the same formated timestamp as the clamav process
type logWriter struct{}

func (writer logWriter) Write(bytes []byte) (int, error) {
	return fmt.Printf("%v -> %v", time.Now().UTC().Format(time.ANSIC), string(bytes))
}
