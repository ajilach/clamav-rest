package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dutchcoders/go-clamd"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var opts map[string]string

func init() {
	log.SetOutput(ioutil.Discard)
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
			version_values := strings.Split(strings.Replace(version_string.Raw, "ClamAV ", "", 1),"/")
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

//This is where the action happens.
func scanHandler(w http.ResponseWriter, r *http.Request) {
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
		for {
			part, err := reader.NextPart()
			if err == io.EOF {
				break
			}

			//if part.FileName() is empty, skip this iteration.
			if part.FileName() == "" {
				continue
			}

			fmt.Printf(time.Now().Format(time.RFC3339) + " Started scanning: " + part.FileName() + "\n")
			var abort chan bool
			response, err := c.ScanStream(part, abort)
			for s := range response {
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				respJson := fmt.Sprintf("{ \"Status\": \"%s\", \"Description\": \"%s\" }", s.Status, s.Description)
				switch s.Status {
				case clamd.RES_OK:
					w.WriteHeader(http.StatusOK)
				case clamd.RES_FOUND:
					w.WriteHeader(http.StatusNotAcceptable)
				case clamd.RES_ERROR:
					w.WriteHeader(http.StatusBadRequest)
				case clamd.RES_PARSE_ERROR:
					w.WriteHeader(http.StatusPreconditionFailed)
				default:
					w.WriteHeader(http.StatusNotImplemented)
				}
				fmt.Fprint(w, respJson)
				fmt.Printf(time.Now().Format(time.RFC3339)+" Scan result for: %v, %v\n", part.FileName(), s)
			}
			fmt.Printf(time.Now().Format(time.RFC3339) + " Finished scanning: " + part.FileName() + "\n")
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func scanHandlerBody(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	c := clamd.NewClamd(opts["CLAMD_PORT"])

	fmt.Printf(time.Now().Format(time.RFC3339) + " Started scanning plain body\n")
	var abort chan bool
	defer r.Body.Close()
	response, _ := c.ScanStream(r.Body, abort)
	for s := range response {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		respJson := fmt.Sprintf("{ Status: %q, Description: %q }", s.Status, s.Description)
		switch s.Status {
		case clamd.RES_OK:
			w.WriteHeader(http.StatusOK)
		case clamd.RES_FOUND:
			w.WriteHeader(http.StatusNotAcceptable)
		case clamd.RES_ERROR:
			w.WriteHeader(http.StatusBadRequest)
		case clamd.RES_PARSE_ERROR:
			w.WriteHeader(http.StatusPreconditionFailed)
		default:
			w.WriteHeader(http.StatusNotImplemented)
		}
		fmt.Fprint(w, respJson)
		fmt.Printf(time.Now().Format(time.RFC3339)+" Scan result for plain body: %v\n", s)
	}
}

func waitForClamD(port string, times int) {
	clamdTest := clamd.NewClamd(port)
	clamdTest.Ping()
	version, err := clamdTest.Version()

	if err != nil {
		if times < 30 {
			fmt.Printf("clamD not running, waiting times [%v]\n", times)
			time.Sleep(time.Second * 4)
			waitForClamD(port, times+1)
		} else {
			fmt.Printf("Error getting clamd version: %v\n", err)
			os.Exit(1)
		}
	} else {
	  	for version_string := range version {
			fmt.Printf("Clamd version: %#v\n", version_string.Raw)
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

	http.HandleFunc("/scan", scanHandler)
	http.HandleFunc("/scanPath", scanPathHandler)
        http.HandleFunc("/version", clamversion)
	http.HandleFunc("/", home)

	// Prometheus metrics
	http.Handle("/metrics", promhttp.HandlerFor(
		reg,
		promhttp.HandlerOpts{
			// Opt into OpenMetrics to support exemplars.
			EnableOpenMetrics: true,
		},
	))

	// Start the HTTPS server in a goroutine
	go http.ListenAndServeTLS(fmt.Sprintf(":%s", opts["SSL_PORT"]), "/etc/ssl/clamav-rest/server.crt", "/etc/ssl/clamav-rest/server.key", nil)

	// Start the HTTP server
	http.ListenAndServe(fmt.Sprintf(":%s", opts["PORT"]), nil)
}
