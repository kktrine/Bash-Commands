package post_run_command

import (
	"bash-commands/server/command_result"
	"bash-commands/server/reqid"
	"encoding/json"
	"log/slog"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
)

type Response struct {
	Error         string `json:"error,omitempty"`
	PID           int    `json:"pid,omitempty"`
	ID            int64  `json:"id,omitempty"`
	Output        string `json:"output,omitempty"`
	CommandErrors string `json:"command_errors,omitempty"`
}

type CommandRunner interface {
	FindAndRun(id int64) (*exec.Cmd, error)
	Start(cmd *exec.Cmd) (int, error)
	Exec(cmd *exec.Cmd, pid int, id int64) (*command_result.CommandResult, error)
}

func NewRunner(log *slog.Logger, runner CommandRunner) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		const op = "handlers.url.run_command.NewRunner"
		logger := log.With(
			slog.String("op", op),
			slog.String("request_id", reqid.GetRequestId(r.Context())),
		)

		id, err := strconv.ParseInt(strings.TrimPrefix(r.URL.Path, "/"), 10, 64)
		if err != nil {
			logger.Error("invalid id (path) " + "error: " + err.Error())
			http.Error(w, "invalid path", http.StatusNotFound)
			return
		}
		cmd, err := runner.FindAndRun(id)
		if err != nil {
			logger.Error("can't run command " + "error: " + err.Error())
			http.Error(w, "can't run command or command not exists", http.StatusInternalServerError)
			return
		}
		pid, err := runner.Start(cmd)
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
		res, err := runner.Exec(cmd, pid, id)
		if err != nil {

			json.NewEncoder(w).Encode(&Response{Error: "failed to run command, maybe process was killed"})
			return
		}
		json.NewEncoder(w).Encode(&Response{
			Output:        res.Output,
			CommandErrors: res.CommandErrors,
		})
		return
	}
}
