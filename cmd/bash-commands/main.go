package main

import (
	"bash-commands/internal/config"
	"bash-commands/internal/logs"
)

func main() {
	cfg := config.MustLoad()
	log := logs.SetupLogger(cfg.LogFilePath)
	log.Info("Starting server at", cfg.HTTPServer.Address)

}
