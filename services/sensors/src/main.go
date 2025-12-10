package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Log request path
		log.Printf("Path: %s\n", r.URL.Path)

		// Log headers
		log.Println("Headers:")
		for name, values := range r.Header {
			for _, v := range values {
				log.Printf("  %s: %s\n", name, v)
			}
		}

		// Read and log body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading body: %v", err)
			http.Error(w, "could not read body", http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		log.Println("Body:")
		log.Println(string(body))

		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	log.Println("Listening on :80")
	log.Fatal(http.ListenAndServe(":80", nil))
}
