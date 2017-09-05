package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dutchcoders/go-clamd"
)

var opts map[string]string

func init() {
	log.SetOutput(ioutil.Discard)
}

func home(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "...running...")
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
				respJson := fmt.Sprintf("{ Status: \"%s\", Description: \"%s\" }", s.Status, s.Description)
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
	http.HandleFunc("/", home)

	//Listen on port PORT
	if opts["PORT"] == "" {
		opts["PORT"] = "9000"
	}
	fmt.Printf("Listening on port " + opts["PORT"])
	http.ListenAndServe(":"+opts["PORT"], nil)
}
