package logs

import (
	"log/slog"
	"os"
)

func SetupLogger(path string) *slog.Logger {
	file, err := os.Open(path)
	if err != nil {
		panic("Can't setup logger: " + err.Error())
	}
	log := slog.New(slog.NewTextHandler(file, &slog.HandlerOptions{Level: slog.LevelInfo}))
	return log
}
