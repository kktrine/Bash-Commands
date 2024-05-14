package get_one_command

import (
	"bash-commands/server/reqid"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
)

type Response struct {
	Error   string `json:"error,omitempty"`
	Command string `json:"command,omitempty"`
}

type CommandGetter interface {
	Get(id int64) (string, error)
}

func NewGetter(log *slog.Logger, getter CommandGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		const op = "handlers.url.stop.NewDeleter"
		logger := log.With(
			slog.String("op", op),
			slog.String("request_id", reqid.GetRequestId(r.Context())),
		)

		id, err := strconv.ParseInt(strings.TrimPrefix(r.URL.Path, "/"), 10, 64)
		if err != nil {
			logger.Error("invalid id (path) " + "error: " + err.Error())
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(&Response{Error: "invalid path"})
			return
		}
		res, err := getter.Get(id)
		if err != nil {
			logger.Error("error: " + err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(&Response{Error: "error"})
			return
		}
		if res == "" {
			logger.Error("can't find id error: " + err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(&Response{Error: "no command with given id"})
			return
		}
		json.NewEncoder(w).Encode(&Response{Command: res})
		return
	}
}
