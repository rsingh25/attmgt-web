package main

import (
	"attmgt-web/internal/database"
	"attmgt-web/internal/logger"
	"attmgt-web/internal/server"
	"log/slog"
	"net/http"
	"time"

	_ "github.com/joho/godotenv/autoload" //autoloads .env
)

var log *slog.Logger

func init() {
	log = logger.Logger.With("package", "main")
}

func main() {

	var httpPort string
	var db database.Service

	env := logger.GetenvStr("APP_ENV", "")
	if env == "local" {
		httpPort = "8080"
		db = database.NewService(true, false)
	} else {
		httpPort = "80"
		db = database.NewService(true, true)
	}
	defer db.Close()

	handler := server.NewServer(db)

	httpServer := &http.Server{
		Addr:         ":" + httpPort,
		Handler:      handler,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	log.Info("Staring HTTP server", "port", httpPort)
	err := httpServer.ListenAndServe()
	if err != nil {
		log.Error("Server failed to start", "err", err)
	}
}
