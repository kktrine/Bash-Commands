package post_new_command

import (
	"bash-commands/internal/storage/storageErrors"
	"bash-commands/server/command_result"
	"bash-commands/server/reqid"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os/exec"
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
	AddAndRun(command string) (int64, *exec.Cmd, error)
	Start(cmd *exec.Cmd) (int, error)
	Exec(cmd *exec.Cmd, pid int, id int64) (*command_result.CommandResult, error)
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
			http.Error(w, "failed to parse request", http.StatusBadRequest)
			return
		}

		id, cmd, err := poster.AddAndRun(req.Command)

		if errors.Is(err, storageErrors.ErrDuplicateEntry) {
			logger.Error("failed to post_new_command command + " + "error: " + err.Error())
			http.Error(w, "command already exists", http.StatusBadRequest)
			return
		}
		if err != nil {
			logger.Error("failed to post_new_command command " + "error: " + err.Error())
			http.Error(w, "failed to post_new_command", http.StatusInternalServerError)
			return
		}
		pid, err := poster.Start(cmd)
		if err != nil {
			logger.Error("failed to run command", "error", err)
			http.Error(w, "failed to run command", http.StatusInternalServerError)
			return
		}
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(&Response{ID: id, PID: pid})
		flusher.Flush()
		res, err := poster.Exec(cmd, pid, id)
		if err != nil {
			logger.Error("failed to exec command", "error", err)
			json.NewEncoder(w).Encode(&Response{Error: "failed to run command, maybe process was killed"})
			return
		}
		json.NewEncoder(w).Encode(res)
		return
	}
}
