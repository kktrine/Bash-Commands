package storage

import (
	"bash-commands/internal/storage/postgresql"
)

type Storage struct {
	db *postgresql.Postgres
}

func New(db string) *Storage {
	st := Storage{}
	st.db = postgresql.NewPostgresRepository(db)
	return &st
}

func (s Storage) Post(command string) (int64, error) {
	return s.db.InsertCommand(command)

}

func (s Storage) Run(command string) (int64, error) {
	//TODO: impl
	return 0, nil
}

func (s Storage) Stop() error {
	return s.db.Stop()
}
