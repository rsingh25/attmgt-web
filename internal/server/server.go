package server

import (
	"attmgt-web/internal/logger"
	"html/template"
	"log/slog"
	"net/http"

	"attmgt-web/internal/database"
	"attmgt-web/internal/util"
)

type Server struct {
	port          int
	env           string
	db            database.Service
	templateCache map[string]*template.Template
}

var log *slog.Logger

func init() {
	log = logger.Logger.With("package", "server")
}

func NewServer(db database.Service) http.Handler {
	NewServer := &Server{
		port:          util.GetenvInt("PORT", 8080),
		db:            db,
		templateCache: make(map[string]*template.Template),
	}

	return NewServer.RegisterRoutes()
}
