package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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

func init() {
	log.SetOutput(io.Discard)
}

func clamversion(w http.ResponseWriter, r *http.Request) {
	c := clamd.NewClamd(opts["CLAMD_PORT"])

	version, err := c.Version()

	if err != nil {
		errJson, eErr := json.Marshal(err)
		if eErr != nil {
			fmt.Println(eErr)
			return
		}
		fmt.Fprint(w, string(errJson))
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	for version_string := range version {
		if strings.HasPrefix(version_string.Raw, "ClamAV ") {
			version_values := strings.Split(strings.Replace(version_string.Raw, "ClamAV ", "", 1), "/")
			respJson := fmt.Sprintf("{ \"Clamav\": \"%s\" }", version_values[0])
			if len(version_values) == 3 {
				respJson = fmt.Sprintf("{ \"Clamav\": \"%s\", \"Signature\": \"%s\" , \"Signature_date\": \"%s\" }", version_values[0], version_values[1], version_values[2])
			}
			fmt.Fprint(w, string(respJson))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	c := clamd.NewClamd(opts["CLAMD_PORT"])

	response, err := c.Stats()

	if err != nil {
		errJson, eErr := json.Marshal(err)
		if eErr != nil {
			fmt.Println(eErr)
			return
		}
		fmt.Fprint(w, string(errJson))
		return
	}

	resJson, eRes := json.Marshal(response)
	if eRes != nil {
		fmt.Println(eRes)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprint(w, string(resJson))
}

func scanPathHandler(w http.ResponseWriter, r *http.Request) {
	paths, ok := r.URL.Query()["path"]
	if !ok || len(paths[0]) < 1 {
		log.Println("Url Param 'path' is missing")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("URL param 'path' is missing"))
		return
	}

	path := paths[0]

	c := clamd.NewClamd(opts["CLAMD_PORT"])
	response, err := c.AllMatchScanFile(path)

	if err != nil {
		errJson, eErr := json.Marshal(err)
		if eErr != nil {
			fmt.Println(eErr)
			return
		}
		fmt.Fprint(w, string(errJson))
		return
	}

	var scanResults []*clamd.ScanResult
	for responseItem := range response {
		if responseItem.Status == clamd.RES_FOUND {
			noOfFoundViruses.Inc()
		}
		scanResults = append(scanResults, responseItem)
	}

	resJson, eRes := json.Marshal(scanResults)
	if eRes != nil {
		fmt.Println(eRes)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprint(w, string(resJson))
}

func v2ScanHandler(w http.ResponseWriter, r *http.Request) {
	scanner(w, r, 2)
}

// old endpoint version, set deprecation header to indicate usage of the new /v2/scan
func scanHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Deprecation", "version=v1")
	v2url := fmt.Sprintf("%s%s/v2/scan", string(r.URL.Scheme), r.Host)
	w.Header().Add("Link", fmt.Sprintf("%v; rel=successor-version", v2url))

	scanner(w, r, 1)
}

// This is where the action happens.
func scanner(w http.ResponseWriter, r *http.Request, version int) {
	switch r.Method {
	//POST takes the uploaded file(s) and saves it to disk.
	case "POST":
		c := clamd.NewClamd(opts["CLAMD_PORT"])
		//get the multipart reader for the request.
		reader, err := r.MultipartReader()

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		//copy each part to destination.
		resp := []scanResponse{}
		for {
			part, err := reader.NextPart()
			if err == io.EOF {
				break
			}

			//if part.FileName() is empty, skip this iteration.
			if part.FileName() == "" {
				if version == 2 {
					fileResp := scanResponse{Status: "ERROR", Description: "MimePart FileName missing", httpStatus: 422}
					resp = append(resp, fileResp)
					fmt.Printf("%v Not scanning, MimePart FileName not supplied\n", time.Now().Format(time.RFC3339))
				}
				continue
			}

			fmt.Printf("%v Started scanning: %v\n", time.Now().Format(time.RFC3339), part.FileName())
			var abort chan bool
			response, err := c.ScanStream(part, abort)
			if err != nil {
				//error occurred, response is nil, create a custom response and send it on the channel to handle it together with the other errors.
				response = make(chan *clamd.ScanResult)
				scanErrResult := &clamd.ScanResult{Status: clamd.RES_PARSE_ERROR, Description: "File size limit exceeded"}
				go func() {
					response <- scanErrResult
					close(response)
					fmt.Printf("%v Clamd returned an error, probably a too large file as input (causing broken pipe and closed connection) %v\n", time.Now().Format(time.RFC3339), err)
					//The underlying service closes the connection if the file is to large, logging output
					//We never receive the clamd output of `^INSTREAM: Size limit reached` up here, just a closed connection.
				}()

			}
			for s := range response {
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				eachResp := scanResponse{Status: s.Status, Description: s.Description}
				if version == 2 {
					eachResp.FileName = part.FileName()
					fmt.Printf("%v Scanned file %v\n", time.Now().Format(time.RFC3339), part.FileName())
				}
				//Set each possible status and then send the most appropriate one
				eachResp.httpStatus = getHttpStatusByClamStatus(s)
				resp = append(resp, eachResp)
				fmt.Printf("%v Scan result for: %v, %v\n", time.Now().Format(time.RFC3339), part.FileName(), s)
			}
			fmt.Printf("%v Finished scanning: %v\n", time.Now().Format(time.RFC3339), part.FileName())
		}
		w.WriteHeader(getResponseStatus(resp))
		if version == 2 {
			jsonRes, jErr := json.Marshal(resp)
			if jErr != nil {
				fmt.Printf("%v Error marshalling json, %v\n", time.Now().Format(time.RFC3339), jErr)
			}
			fmt.Fprint(w, string(jsonRes))
		} else {
			for _, v := range resp {
				jsonRes, jErr := json.Marshal(v)
				if jErr != nil {
					fmt.Printf("%v Error marshalling json, %v\n", time.Now().Format(time.RFC3339), jErr)
				}
				fmt.Fprint(w, string(jsonRes))
			}
		}

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func getHttpStatusByClamStatus(result *clamd.ScanResult) int {
	switch result.Status {
	case clamd.RES_OK:
		return http.StatusOK //200
	case clamd.RES_FOUND:
		fmt.Printf("%v Virus FOUND\n", time.Now().Format(time.RFC3339))
		return http.StatusNotAcceptable //406
	case clamd.RES_ERROR:
		return http.StatusBadRequest //400
	case clamd.RES_PARSE_ERROR:
		if result.Description == "File size limit exceeded" {
			return http.StatusRequestEntityTooLarge //413
		} else {
			return http.StatusPreconditionFailed //412
		}
	default:
		return http.StatusNotImplemented //501
	}
}

// this func returns 406 if one file contains a virus
func getResponseStatus(responses []scanResponse) int {
	result := 200
	for _, r := range responses {
		switch r.httpStatus {
		case 406:
			//uptick the prometheus counter for detected viruses.
			noOfFoundViruses.Inc()
			//early return if virus is found
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

	fmt.Printf("%v Started scanning plain body\n", time.Now().Format(time.RFC3339))
	var abort chan bool
	defer r.Body.Close()
	response, err := c.ScanStream(r.Body, abort)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err != nil {
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		resp := scanResponse{Status: clamd.RES_PARSE_ERROR, Description: "File size limit exceeded"}
		fmt.Printf("%v Clamd returned error, broken pipe and closed connection can indicate too large file, %v\n", time.Now().Format(time.RFC3339), err)
		jsonResp, jErr := json.Marshal(resp)
		if jErr != nil {
			fmt.Printf("%v Error marshalling json, %v\n", time.Now().Format(time.RFC3339), jErr)
		}
		fmt.Fprint(w, string(jsonResp))
		return
	}
	for s := range response {

		resp := scanResponse{Status: s.Status, Description: s.Description}
		//respJson := fmt.Sprintf("{ Status: %q, Description: %q }", s.Status, s.Description)
		resp.httpStatus = getHttpStatusByClamStatus(s)

		resps := []scanResponse{}
		resps = append(resps, resp)
		w.WriteHeader(getResponseStatus(resps))
		fmt.Fprint(w, resp)
		fmt.Printf("%v Scan result for plain body: %v\n", time.Now().Format(time.RFC3339), s)
	}
}

func waitForClamD(port string, times int) {
	clamdTest := clamd.NewClamd(port)
	clamdTest.Ping()
	version, err := clamdTest.Version()

	if err != nil {
		if times < 30 {
			fmt.Printf("%v clamD not running, waiting times [%v]\n", time.Now().Format(time.RFC3339), times)
			time.Sleep(time.Second * 4)
			waitForClamD(port, times+1)
		} else {
			fmt.Printf("%v Error getting clamd version: %v\n", time.Now().Format(time.RFC3339), err)
			os.Exit(1)
		}
	} else {
		for version_string := range version {
			fmt.Printf("%v Clamd version: %#v\n", time.Now().Format(time.RFC3339), version_string.Raw)
		}
	}
}

func main() {

	opts = make(map[string]string)

	// https://github.com/prometheus/client_golang/blob/main/examples/gocollector/main.go
	reg := prometheus.NewRegistry()
	reg.MustRegister(collectors.NewBuildInfoCollector())
	reg.MustRegister(collectors.NewGoCollector(
		collectors.WithGoCollections(collectors.GoRuntimeMemStatsCollection | collectors.GoRuntimeMetricsCollection),
	))
	reg.MustRegister(noOfFoundViruses)

	for _, e := range os.Environ() {
		pair := strings.Split(e, "=")
		opts[pair[0]] = pair[1]
	}

	if opts["CLAMD_PORT"] == "" {
		opts["CLAMD_PORT"] = "tcp://localhost:3310"
	}

	fmt.Printf("Starting clamav rest bridge\n")
	fmt.Printf("Connecting to clamd on %v\n", opts["CLAMD_PORT"])
	waitForClamD(opts["CLAMD_PORT"], 1)

	fmt.Printf("Connected to clamd on %v\n", opts["CLAMD_PORT"])
	mux := http.NewServeMux()
	//Add cors middleware
	c := cors.New(getCorsPolicy())

	mux.HandleFunc("POST /scan", scanHandler)
	mux.HandleFunc("POST /v2/scan", v2ScanHandler)
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
	//Attach the cors middleware to the middleware chain/request pipeline
	handler := c.Handler(mux)
	// Start the HTTPS server in a goroutine
	go func() {
		log.Fatal(http.ListenAndServeTLS(fmt.Sprintf(":%s", opts["SSL_PORT"]), "/etc/ssl/clamav-rest/server.crt", "/etc/ssl/clamav-rest/server.key", handler))
	}()
	// Start the HTTP server
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", opts["PORT"]), handler))
}

func getCorsPolicy() cors.Options {
	envs := os.Environ()
	var allow_origins []string

	for _, env := range envs {
		e := strings.Split(env, "=")
		if strings.EqualFold(e[0], "allow_origins") {
			allow_origins = strings.Split(e[1], ";")
		}
	}
	if len(allow_origins) == 0 {
		allow_origins = []string{"*"}
	}
	return cors.Options{
		AllowedOrigins:   allow_origins,
		AllowedHeaders:   []string{"Content-Type", "Accept-Language", "Accept"},
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodOptions},
		AllowCredentials: false,
	}
}
