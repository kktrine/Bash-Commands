package main

import (
	"bash-commands/internal/config"
	"bash-commands/internal/logs"
	"bash-commands/internal/storage"
	"bash-commands/server"
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	cfg := config.MustLoad()
	log := logs.SetupLogger(cfg.LogFilePath)
	st := storage.New(cfg.Postgres)
	srv := server.NewServer(log, st)
	go srv.Start(cfg.Address)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	//test := post.NewPoster(log, st)
	//_ = test
	// Ожидание сигнала завершения
	<-done

	// Создание контекста с таймаутом для завершения работы сервера
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Остановка сервера
	if err := srv.Stop(ctx); err != nil {
		log.Error("error stopping server: %v\n", err)
		return
	}
	log.Info("server stopped")
}
