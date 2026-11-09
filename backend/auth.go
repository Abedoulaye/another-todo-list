package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

func register(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	var u User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		clientError(w, http.StatusBadRequest, "invalid request body at register")
		return
	}
	u.Email = strings.ToLower(strings.TrimSpace(u.Email))
	if err := validateEmail(u.Email); err != nil {
		clientError(w, http.StatusBadRequest, "invalid email")
		return
	}
	if err := validatePassword(u.Password); err != nil {
		clientError(w, http.StatusBadRequest, "invalid password")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		serverError(w, err, "failed bcrypt hash")
		return
	}

	if err := db.QueryRow(ctx, "INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id", u.Email, string(hashedPassword)).Scan(&u.ID); err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			clientError(w, http.StatusConflict, "email already exists")
			return
		}
		serverError(w, err, "register insert")
		return
	}

	u.Password = ""

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(u)
}

func login(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	var u User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		clientError(w, http.StatusBadRequest, "invalid request body at login")
		return
	}

	u.Email = strings.ToLower(strings.TrimSpace(u.Email))
	if u.Email == "" || u.Password == "" {
		clientError(w, http.StatusBadRequest, "invalid email or password")
		return
	}

	var storedHash string
	if err := db.QueryRow(ctx, "SELECT id, password_hash, role FROM users WHERE email = $1", u.Email).Scan(&u.ID, &storedHash, &u.Role); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			clientError(w, http.StatusUnauthorized, "invalid email and password")
			return
		}
		serverError(w, err, "login select query")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(u.Password)); err != nil {
		clientError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	accessToken, err := generateAccessToken(u.ID, u.Role)
	if err != nil {
		serverError(w, err, "generate access token")
		return
	}

	refreshToken, err := generateRefreshToken()
	if err != nil {
		serverError(w, err, "generate refresh token")
		return
	}

	_, err = db.Exec(ctx, "INSERT INTO refresh_tokens (user_id, token_hash, expires) VALUES ($1, $2, $3)", u.ID, hashToken(refreshToken), time.Now().Add(30*24*time.Hour))
	if err != nil {
		serverError(w, err, "insert into refresh tokens table")
		return
	}

	u.Password = ""
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user":          u,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func generateAccessToken(id int, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": id,
		"role":    role,
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

func generateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func validateToken(tokenString string) (int, string, error) {
	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		return 0, "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := int(claims["user_id"].(float64))
		role := claims["role"].(string)
		return userID, role, nil
	}

	return 0, "", fmt.Errorf("invalid token")
}

func refresh(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		clientError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var userID int
	var role string
	if err := db.QueryRow(ctx, "SELECT u.id, u.role FROM refresh_tokens rt JOIN users ON u.id = rt.user_id WHERE rt.token_hash = $1 AND rt.expires > NOW()", hashToken(req.RefreshToken)).Scan(&userID, &role); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			clientError(w, http.StatusUnauthorized, "invalid or expired refresh token")
		}
		serverError(w, err, "refresh select query")
		return
	}

	accessToken, err := generateAccessToken(userID, role)
	if err != nil {
		serverError(w, err, "refresh generate access token")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"access_token": accessToken,
	})

}

func logout(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		clientError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	_, err := db.Exec(ctx, "DELETE FROM refresh_tokens WHERE token_hash = $1", hashToken(req.RefreshToken))
	if err != nil {
		serverError(w, err, "logout delete token")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
