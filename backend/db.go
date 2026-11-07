package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

var db *pgxpool.Pool

func init() {
	if err := godotenv.Load(); err != nil {
		slog.Error("No .env file found")
	}

	config, err := pgxpool.ParseConfig("")
	if err != nil {
		log.Fatalf("Unable to parse database URL: %v", err)
	}

	config.ConnConfig.User = os.Getenv("DB_USER")
	config.ConnConfig.Host = os.Getenv("DB_HOST")

	port, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		slog.Error("Invalid Port: %v", err)
	}
	config.ConnConfig.Port = uint16(port)
	config.ConnConfig.Host = os.Getenv("DB_HOST")
	config.ConnConfig.Database = os.Getenv("DB_NAME")

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatalf("unable to create connnection pool: %v", err)
	}
	db = pool
}
