package delete_one

import (
	"bash-commands/server/reqid"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
)

//go:generate go run github.com/vektra/mockery/v2@v2.43.0 --name=CommandDeleter
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
			http.Error(w, "invalid path", http.StatusBadRequest)
			return
		}
		found, err := deleter.Delete(id)
		if err != nil {
			logger.Error("error: " + err.Error())
			http.Error(w, "error", http.StatusInternalServerError)
			return
		}
		if !found {
			logger.Error("id not found ")
			http.Error(w, "id not found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}
}
