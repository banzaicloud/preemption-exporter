package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

// use this minimal http server to test the exporter locally
// change constant to have metadataEndpoint = "http://metadata.google.internal/computeMetadata/v1/instance/"
func main() {

	http.HandleFunc("/computeMetadata/v1/instance/id", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "in-12345")
	})

	http.HandleFunc("/computeMetadata/v1/instance/preempted", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "TRUE")
	})

	log.Fatal(http.ListenAndServe(":9092", nil))

}
