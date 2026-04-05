package main

import (
	"log/slog"
	"os"
)

func NewLogger() *slog.Logger {
	var handler slog.Handler
	if os.Getenv("ENV") == "production" {
		opts := &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		opts := &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	return slog.New(handler)
}
