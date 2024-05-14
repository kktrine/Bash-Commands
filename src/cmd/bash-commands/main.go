package main

import (
	"bash-commands/internal/logs"
	"bash-commands/internal/storage"
	"bash-commands/server"
	"context"
	"flag"
	"github.com/joho/godotenv"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	docker := flag.Bool("docker", false, "set this if you run program in docker")
	flag.Parse()
	var err error
	if !*docker {
		err = godotenv.Load(".env")
	} else {
		err = godotenv.Load(".env_docker")
	}
	if err != nil {
		panic(err.Error())
	}
	log := logs.SetupLogger(os.Getenv("LOG_PATH"))
	st := storage.New(os.Getenv("POSTGRES"))
	srv := server.NewServer(log, st)
	go srv.Start(os.Getenv("HOST"))

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

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
