package util

import (
	"attmgt-web/internal/logger"
	"log/slog"
	"runtime/debug"
)

var log *slog.Logger

func init() {
	log = logger.Logger.With("package", "main")
}

func Map[T, V any](ts []T, fn func(T) V) []V {
	result := make([]V, len(ts))
	for i, t := range ts {
		result[i] = fn(t)
	}
	return result
}

func Must[T any](t T, err error) T {
	if err != nil {
		log.Error("Failed with error (panicing)", "error", err.Error(), "stack", debug.Stack())
		panic(err) //No recovery
	}
	return t
}
