package utils

import (
	"log/slog"
	"os"
)

// Logger is the structured logger for the application
var Logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
	Level: slog.LevelInfo,
}))

// SetLogLevel changes the log level at runtime
func SetLogLevel(level slog.Level) {
	Logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
}