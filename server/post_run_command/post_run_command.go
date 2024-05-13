package post_run_command

import (
	"bash-commands/server/reqid"
	"encoding/json"
	"log/slog"
	"net/http"
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
	Run(id int64) (*Response, error)
}

func NewRunner(log *slog.Logger, runner CommandRunner) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		const op = "handlers.url.post_new_command.NewPoster"
		logger := log.With(
			slog.String("op", op),
			slog.String("request_id", reqid.GetRequestId(r.Context())),
		)

		id, err := strconv.ParseInt(strings.TrimPrefix(r.URL.Path, "/"), 10, 64)
		if err != nil {
			logger.Error("invalid id (path) " + "error: " + err.Error())
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(&Response{Error: "invalid path"})
			return
		}
		res, err := runner.Run(id)
		if err != nil {
			logger.Error("can't run command " + "error: " + err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(&Response{Error: "can't run command"})
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
		return
	}
}
