package server

import (
	"bash-commands/internal/storage"
	delete_one "bash-commands/server/delete_one"
	"bash-commands/server/get_all_commands"
	"bash-commands/server/post_new_command"
	"bash-commands/server/post_run_command"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Server struct {
	server                *http.Server
	mux                   *http.ServeMux
	logger                *slog.Logger
	st                    *storage.Storage
	postNewCommandHandler http.HandlerFunc
	postRunCommandHandler http.HandlerFunc
	getCommandsHandler    http.HandlerFunc
	deleteHandler         http.HandlerFunc
}

func NewServer(log *slog.Logger, st *storage.Storage) *Server {
	srv := Server{logger: log, mux: http.NewServeMux(), st: st}
	srv.server = &http.Server{Handler: srv.mux}

	srv.postNewCommandHandler = post_new_command.NewPoster(srv.logger, srv.st)
	srv.postRunCommandHandler = post_run_command.NewRunner(srv.logger, srv.st)
	srv.getCommandsHandler = get_all_commands.NewGetter(srv.logger, srv.st)
	srv.deleteHandler = delete_one.NewDeleter(srv.logger, srv.st)

	srv.mux.HandleFunc("/", srv.mainHandler)
	srv.server.Handler = srv.recoverer(srv.server.Handler)
	srv.server.Handler = srv.logRequest(srv.server.Handler)
	srv.server = &http.Server{Handler: srv.server.Handler}

	return &srv
}

func (s Server) mainHandler(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET" && r.URL.Path == "/":
		s.getCommandsHandler(w, r)
	case r.Method == "POST" && r.URL.Path == "/":
		s.postNewCommandHandler(w, r)
	case r.Method == "POST" && strings.HasPrefix(r.URL.Path, "/"):
		s.postRunCommandHandler(w, r)
	case r.Method == "GET" && strings.HasPrefix(r.URL.Path, "/"):
		http.Error(w, "Method Not Implemented", http.StatusNotImplemented)
	case r.Method == "DELETE" && strings.HasPrefix(r.URL.Path, "/"):
		s.deleteHandler(w, r)
	default:
		http.Error(w, "Method Not Exist", http.StatusMethodNotAllowed)
	}
}

func (s Server) Start(addr string) {
	s.server.Addr = addr
	s.logger.Info("starting server at " + addr)
	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.logger.Error("error starting server: %v\n", err)
		panic(err)
	}
}

func (s Server) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := time.Now()
		next.ServeHTTP(w, r)
		s.logger.Info(
			r.Method + " " +
				r.URL.Path + " " +
				r.RemoteAddr + " " +
				r.UserAgent() +
				r.RequestURI +
				time.Since(t).String(),
		)
	})
}

func (s Server) recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				s.logger.Error("Возникла ошибка: %v ", err)
				http.Error(w, "Internal Server Error (recover)", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (s Server) Stop(ctx context.Context) error {
	err := s.st.Stop()
	if err != nil {
		s.logger.Error("error stopping DB connection: " + err.Error())
	}
	return s.server.Shutdown(ctx)
}

// GetHandler - обработчик для GET запросов
func getHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Привет, мир!"))
}

// PostHandler - обработчик для POST запросов
func postHandler(w http.ResponseWriter, r *http.Request) {
	// Обработка POST запроса
	w.Write([]byte("POST запрос успешно обработан!"))
}

// DeleteHandler - обработчик для DELETE запросов
func deleteHandler(w http.ResponseWriter, r *http.Request) {
	// Обработка DELETE запроса
	id := strings.TrimPrefix(r.URL.Path, "/")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	_ = idInt
	w.Write([]byte("DELETE запрос успешно обработан! " + id))
}
