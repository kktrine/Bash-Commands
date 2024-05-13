package post_new_command

import (
	"bash-commands/internal/storage/storageErrors"
	"bash-commands/server/reqid"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

type Request struct {
	Command string `json:"command"`
}

type Response struct {
	Error         string `json:"error,omitempty"`
	PID           int    `json:"pid,omitempty"`
	ID            int64  `json:"id,omitempty"`
	Output        string `json:"output,omitempty"`
	CommandErrors string `json:"command_errors,omitempty"`
}

type CommandPoster interface {
	Post(command string) (*Response, error)
}

func NewPoster(log *slog.Logger, poster CommandPoster) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		const op = "handlers.url.post_new_command.NewPoster"
		logger := log.With(
			slog.String("op", op),
			slog.String("request_id", reqid.GetRequestId(r.Context())),
		)

		var req Request
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			logger.Error("failed to parse request " + "error: " + err.Error())
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(&Response{Error: "failed to parse request"})
			return
		}

		res, err := poster.Post(req.Command)
		if errors.Is(err, storageErrors.ErrDuplicateEntry) {
			logger.Error("failed to post_new_command command + " + "error: " + err.Error())
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(&Response{Error: "command already exists"})
			return
		}
		if err != nil {
			logger.Error("failed to post_new_command command " + "error: " + err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(&Response{Error: "failed to post_new_command command"})
			return
		}

		if err != nil {
			logger.Error("failed to run command", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(&Response{Error: "failed to run command"})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
		return
	}
}
