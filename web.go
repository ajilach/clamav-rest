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
)

func init() {
	log.SetOutput(ioutil.Discard)
}

//This is where the action happens.
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	//POST takes the uploaded file(s) and saves it to disk.
	case "POST":
		c := clamd.NewClamd("tcp://localhost:3310")
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

	fmt.Printf("Starting clamav rest bridge\n")

	http.HandleFunc("/scan", uploadHandler)

	//static file handler.
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))

	//Listen on port 8080
	http.ListenAndServe(":3030", nil)
}
