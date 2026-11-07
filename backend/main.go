package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	mux := http.NewServeMux()
	srv := &http.Server{
		Handler: mux,
		Addr:    ":" + os.Getenv("ADDRESS"),
	}

	log.Fatal(srv.ListenAndServe())

}
