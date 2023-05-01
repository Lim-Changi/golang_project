package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	// Define a handler function for incoming requests
	handler := func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.RemoteAddr)
		_, err := fmt.Fprintln(w, "Welcome to my Go app server!")
		if err != nil {
			log.Fatal("Response Error!")
		}
	}

	// Set up a server to listen on port 8080
	http.HandleFunc("/", handler)
	log.Println("Starting server on port 8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
