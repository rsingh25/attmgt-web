package logger

import (
	"log/slog"
	"os"
)

var Logger *slog.Logger

func init() {
	logHandleOpts := &slog.HandlerOptions{
		Level: slog.LevelInfo, // Get from env
	}
	defaultAttrs := []slog.Attr{
		slog.String("env", GetenvStr("APP_ENV", "dev")),
	}

	Logger = slog.New(slog.NewJSONHandler(os.Stdout, logHandleOpts).WithAttrs(defaultAttrs))
	slog.SetDefault(Logger)
}

func GetenvStr(key string, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}
