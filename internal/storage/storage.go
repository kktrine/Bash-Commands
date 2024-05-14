package storage

import (
	"bash-commands/internal/storage/postgresql"
	"bash-commands/server/post_new_command"
	"bash-commands/server/post_run_command"
)

type Storage struct {
	db          *postgresql.Postgres
	runningProc map[int64]interface{}
}

func New(db string) *Storage {
	st := Storage{}
	st.db = postgresql.NewPostgresRepository(db)
	return &st
}

func (s Storage) Post(command string) (*post_new_command.Response, error) {
	id, err := s.db.InsertCommand(command)
	if err != nil {
		return nil, err
	}
	res, err := s.db.ExecCommand(id, command)
	if err != nil {
		return nil, err
	}
	return &post_new_command.Response{
		PID:           res.Pid,
		ID:            res.Id,
		Output:        res.Output,
		CommandErrors: res.OutputErrors,
	}, nil
}

func (s Storage) Run(id int64) (*post_run_command.Response, error) {
	res, err := s.db.RunCommandById(id)
	if err != nil {
		return nil, err
	}
	return &post_run_command.Response{
		PID:           res.Pid,
		ID:            res.Id,
		Output:        res.Output,
		CommandErrors: res.OutputErrors,
	}, nil
}

func (s Storage) Stop() error {
	return s.db.Stop()
}
