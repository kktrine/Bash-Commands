package stop_process

import (
	"bash-commands/server/reqid"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
)

type Response struct {
	Error string `json:"error,omitempty"`
}

type CommandStopper interface {
	Kill(pid int) error
}

func NewStopper(log *slog.Logger, stopper CommandStopper) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		const op = "handlers.url.stop.NewStopper"
		logger := log.With(
			slog.String("op", op),
			slog.String("request_id", reqid.GetRequestId(r.Context())),
		)

		pid, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/stop/"))
		if err != nil {
			logger.Error("invalid id (path) " + "error: " + err.Error())
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(&Response{Error: "invalid path"})
			return
		}
		err = stopper.Kill(pid)
		if err != nil {
			logger.Error("error: " + err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(&Response{Error: "pid not found"})
			return
		}

		w.WriteHeader(http.StatusOK)
		return
	}
}
