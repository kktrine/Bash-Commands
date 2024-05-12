package logs

import (
	"log/slog"
	"os"
)

func SetupLogger(path string) *slog.Logger {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic("Can't setup logger: " + err.Error())
	}
	log := slog.New(slog.NewTextHandler(file, &slog.HandlerOptions{Level: slog.LevelInfo}))
	return log
}
