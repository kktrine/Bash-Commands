package main

import (
	"bash-commands/internal/config"
	"bash-commands/internal/logs"
	"bash-commands/internal/mwlogger"
	"bash-commands/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg := config.MustLoad()
	log := logs.SetupLogger(cfg.LogFilePath)
	log.Info("Starting server at", cfg.HTTPServer.Address)
	storage := storage.New(cfg.Postgres)
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	router.Use(mwlogger.New(log))

	router.Route("/command", func(r chi.Router) {
		// TODO: add POST
		//r.Post("/", save.New(log, storage))
		// TODO: add DELETE /url/{id}
	})
}
