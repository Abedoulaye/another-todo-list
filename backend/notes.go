package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

func listNotes(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	userID := r.Context().Value("userID").(int)
	rows, err := db.Query(ctx, "SELECT id, title, body, created_at FROM notes WHERE user_id = $1", userID)
	if err != nil {
		serverError(w, err, "listNotes select query")
		return
	}
	defer rows.Close()

	var notes = []Note{}
	for rows.Next() {
		var note Note
		if err := rows.Scan(&note.ID, &note.Title, &note.Body, &note.CreatedAt); err != nil {
			serverError(w, err, "listNotes scan")
			return
		}
		notes = append(notes, note)

	}

	if err := rows.Err(); err != nil {
		serverError(w, err, "listNotes rows error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notes)
}

func getNote(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	id := r.PathValue("id")

	var note Note
	if err := db.QueryRow(ctx, "SELECT id, title, body, created_at FROM notes WHERE id = $1", id).Scan(&note.ID, &note.Title, &note.Body, &note.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			clientError(w, http.StatusNotFound, "note not found")
			return
		}
		serverError(w, err, "getNote query")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(note)
}

func createNote(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	var note Note
	if err := json.NewDecoder(r.Body).Decode(&note); err != nil {
		clientError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	note.Title = strings.TrimSpace(note.Title)
	note.Body = strings.TrimSpace(note.Body)

	if errors := validateNote(note); len(errors) > 0 {
		clientError(w, http.StatusBadRequest, strings.Join(errors, ", "))
		return
	}

	userID := r.Context().Value("userID").(int)
	if err := db.QueryRow(ctx, "INSERT INTO notes (user_id, title, body, created_at) VALUES ($1, $2, $3, $4) RETURNING id", userID, note.Title, note.Body, time.Now()).Scan(&note.ID); err != nil {
		serverError(w, err, "createNote insert")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(note)
}

func updateNote(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	id := r.PathValue("id")

	var updatedNote Note
	if err := json.NewDecoder(r.Body).Decode(&updatedNote); err != nil {
		clientError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	updatedNote.Title = strings.TrimSpace(updatedNote.Title)
	updatedNote.Body = strings.TrimSpace(updatedNote.Body)

	if errors := validateNote(updatedNote); len(errors) > 0 {
		clientError(w, http.StatusBadRequest, strings.Join(errors, ", "))
		return
	}

	if err := db.QueryRow(ctx, "UPDATE notes SET title = $1, body = $2 WHERE id = $3 RETURNING id", updatedNote.Title, updatedNote.Body, id).Scan(&updatedNote.ID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			clientError(w, http.StatusNotFound, "note not found")
			return
		}
		serverError(w, err, "updateNote db query")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedNote)
}

func deleteNote(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	id := r.PathValue("id")

	result, err := db.Exec(ctx, "DELETE FROM notes WHERE id = $1", id)
	if err != nil {
		serverError(w, err, "deleteNote db query")
		return
	}

	if result.RowsAffected() == 0 {
		clientError(w, http.StatusBadRequest, "note not found")
		return
	}
	w.WriteHeader(http.StatusOK)
}
