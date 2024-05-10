package save

import (
	"bash-commands/internal/storage"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"net/http"
)

type Request struct {
	Command string `json:"url"`
}

type Response struct {
	Error string `json:"error,omitempty"`
	PID   int    `json:"pid,omitempty"`
	ID    int    `json:"id,omitempty"`
}

type CommandSaver interface {
	Save(command string) (int, error)
	Run(command string) (int, error)
}

func NewSaver(log *slog.Logger, saver CommandSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//TODO impl
		w.WriteHeader(http.StatusNotImplemented)
		return

		const op = "handlers.url.save.NewSaver"
		logger := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			logger.Error("failed to parse request", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(&Response{Error: "failed to parse request"})
			return
		}

		id, err := saver.Save(req.Command)
		if errors.Is(err, storage.ErrCommandExists) {
			logger.Error("failed to save command", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(&Response{Error: "command already exists"})
			return
		}
		if err != nil {
			logger.Error("failed to save command", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(&Response{Error: "failed to save command"})
			return
		}

		pid, err := saver.Run(req.Command)
		if err != nil {
			logger.Error("failed to run command", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(&Response{Error: "failed to run command"})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(&Response{
			PID: pid,
			ID:  id,
		})
		return
	}
}
