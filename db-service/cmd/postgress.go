package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/subosito/gotenv"
)

func connectToDB(logger *slog.Logger) *sqlx.DB {
	err := gotenv.Load()
	if err != nil {
		logger.Error("Error get ENV", slog.String("Error", err.Error()))
		os.Exit(1)
	}

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	dns := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := sqlx.Open("postgres", dns)

	logger.Info("Connecting to DB")

	if err != nil {
		logger.Error("Connect to db failed", slog.String("Error", err.Error()))
		os.Exit(1)
	}
	err = db.Ping()
	if err != nil {
		logger.Error("Ping to db failed", slog.String("Error", err.Error()))
		os.Exit(1)
	}
	logger.Info("DB connected successfully")
	return db
}
