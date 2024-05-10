package main

import (
	"bash-commands/internal/config"
	"bash-commands/internal/http-server/save"
	"bash-commands/internal/logs"
	"bash-commands/internal/mwlogger"
	"bash-commands/internal/storage"
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := config.MustLoad()
	log := logs.SetupLogger(cfg.LogFilePath)
	log.Info("Starting server at", cfg.HTTPServer.Address)
	st := storage.New(cfg.Postgres)
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	router.Use(mwlogger.New(log))

	router.Post("/command", save.NewSaver(log, st))
	//TODO: implement get all from bd
	//router.Get("/command", ...)
	//TODO: impl delete
	//router.Delete("/command/{id}", ...)
	// TODO: impl
	//router.Get("/command/{id}")

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Error("failed to start server")
		}
	}()

	log.Info("server started")
	fmt.Println("server started", cfg.HTTPServer.Address)
	<-done
	log.Info("stopping server")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("failed to stop server", err.Error())

		return
	}

	// TODO: close storage

	log.Info("server stopped")
}
