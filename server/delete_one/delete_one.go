package delete_one

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

type CommandDeleter interface {
	Delete(id int64) (bool, error)
}

func NewDeleter(log *slog.Logger, deleter CommandDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		const op = "handlers.url.delete_one.NewDeleter"
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
		found, err := deleter.Delete(id)
		if err != nil {
			logger.Error("error: " + err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(&Response{Error: "error"})
			return
		}
		if !found {
			logger.Error("id not found ")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(&Response{Error: "id not found"})
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}
}
