package post

import (
	"bash-commands/server/reqid"
	"encoding/json"
	"errors"
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

type CommandPoster interface {
	Post(command string) (int, error)
	Run(command string) (int, error)
	ErrCommandExists() error
}

func NewPoster(log *slog.Logger, poster CommandPoster) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		////TODO impl
		//w.WriteHeader(http.StatusNotImplemented)
		//return

		const op = "handlers.url.post.NewPoster"
		logger := log.With(
			slog.String("op", op),
			slog.String("request_id", reqid.GetRequestId(r.Context())),
		)

		var req Request
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			logger.Error("failed to parse request", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(&Response{Error: "failed to parse request"})
			return
		}

		id, err := poster.Post(req.Command)
		if errors.Is(err, poster.ErrCommandExists()) {
			logger.Error("failed to post command", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(&Response{Error: "command already exists"})
			return
		}
		if err != nil {
			logger.Error("failed to post command", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(&Response{Error: "failed to post command"})
			return
		}

		pid, err := poster.Run(req.Command)
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
