package main

import (
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	mux := http.NewServeMux()
	srv := &http.Server{
		Handler: corsMiddleware(mux),
		Addr:    ":" + os.Getenv("ADDRESS"),
	}

	generalLimit := rateLimit(100, time.Minute)
	loginLimit := rateLimit(5, 15*time.Minute)
	registerLimit := rateLimit(3, time.Hour)

	mux.HandleFunc("GET /notes", authMiddleware(generalLimit(listNotes)))
	mux.HandleFunc("GET /notes/{id}", authMiddleware(generalLimit(getNote)))
	mux.HandleFunc("POST /notes", authMiddleware(generalLimit(createNote)))
	mux.HandleFunc("PUT /notes{id}", authMiddleware(generalLimit(updateNote)))
	mux.HandleFunc("DELETE /notes{id}", authMiddleware(generalLimit(deleteNote)))

	mux.HandleFunc("POST /register", registerLimit(register))
	mux.HandleFunc("POST /login", loginLimit(login))
	mux.HandleFunc("POST /refresh", generalLimit(refresh))
	mux.HandleFunc("POST /logout", generalLimit(logout))

	log.Println("server running on port " + srv.Addr)
	log.Fatal(srv.ListenAndServe())

}
