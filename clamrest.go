package main

import (
	"fmt"
	"time"
	"strings"
	"log"
	"io"
	"io/ioutil"
	"net/http"
	"github.com/dutchcoders/go-clamd"
	"os"
)

var opts map[string]string

func init() {
	log.SetOutput(ioutil.Discard)
}

//This is where the action happens.
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	//POST takes the uploaded file(s) and saves it to disk.
	case "POST":
		c := clamd.NewClamd(opts["CLAMD_CONNECTION"])
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
			response,  err := c.ScanStream(part);
			for s := range response {
				if strings.Contains(s, "FOUND") {
					http.Error(w, s, http.StatusInternalServerError)
				} else {
					fmt.Fprintf(w, s)
				}
				fmt.Printf(time.Now().Format(time.RFC3339) + " Scan result for: %v, %v\n", part.FileName(), s)
			}
			fmt.Printf(time.Now().Format(time.RFC3339) + " Finished scanning: " + part.FileName() + "\n")

		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func main() {

	opts = make(map[string]string)

	for _, e := range os.Environ() {
		pair := strings.Split(e, "=")
		opts[pair[0]] = pair[1]
	}

	if opts["CLAMD_CONNECTION"] == "" {
		opts["CLAMD_CONNECTION"] = "tcp://localhost:3031"
	}

	fmt.Printf("Starting clamav rest bridge\n")
	fmt.Printf("Connecting to clamd on %v\n", opts["CLAMD_CONNECTION"])
	clamd_test := clamd.NewClamd(opts["CLAMD_CONNECTION"])
	clamd_test.Ping()
	version, err := clamd_test.Version()
	
	if err != nil {
		fmt.Printf("Error getting clamd version: %v\n", err)
		os.Exit(1)
	}
	for version_string := range version {
		fmt.Printf("Clamd version: %v\n", version_string)
	}
	fmt.Printf("Connected to clamd on %v\n", opts["CLAMD_CONNECTION"])

	http.HandleFunc("/scan", uploadHandler)

	//static file handler.
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))

	//Listen on port 8080
	http.ListenAndServe(":3030", nil)
}
