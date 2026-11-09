package main

import (
	"log"
	"net/http"
)

func clientError(w http.ResponseWriter, status int, message string) {
	log.Printf("Client Error - status: %d : %s ", status, message)
	http.Error(w, message, status)
}
func serverError(w http.ResponseWriter, err error, message string) {
	log.Printf("Server Error - %s: %v", message, err)
	http.Error(w, err.Error(), 500)
}
