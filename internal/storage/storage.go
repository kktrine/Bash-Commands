package storage

import (
	"bash-commands/internal/storage/postgresql"
	"errors"
)

type Storage struct {
	db *postgresql.Postgres
}

func New(db string) *Storage {
	st := Storage{}
	st.db = postgresql.NewPostgresRepository(db)
	return &st
}

func (s Storage) ErrCommandExists() error { return errors.New("command exists") }

func (s Storage) Save(command string) (int, error) {
	//TODO: impl
	return 0, nil
}

func (s Storage) Run(command string) (int, error) {
	//TODO: impl
	return 0, nil
}
