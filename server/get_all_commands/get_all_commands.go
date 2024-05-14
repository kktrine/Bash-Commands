package get_all_commands

import (
	"bash-commands/server/reqid"
	"encoding/json"
	"log/slog"
	"net/http"
)

type Response struct {
	Commands []Command `json:"commands,omitempty"`
	Error    string    `json:"error,omitempty"`
}

type Command struct {
	Id      int64  `json:"id"`
	Command string `json:"command"`
}

type CommandsGetter interface {
	Get() (*Response, error)
}

func NewGetter(log *slog.Logger, getter CommandsGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		const op = "handlers.url.get_all_commands.NewGetter"
		logger := log.With(
			slog.String("op", op),
			slog.String("request_id", reqid.GetRequestId(r.Context())),
		)

		res, err := getter.Get()
		if err != nil {
			logger.Error("failed to get commands " + "error: " + err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(&Response{Error: "failed to get commands "})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
		return
	}
}
