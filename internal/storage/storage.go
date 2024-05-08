package storage

import "bash-commands/internal/storage/postgresql"

type Storage struct {
	db *postgresql.Postgres
	//cache *cashe.Cache
}

func New(db string) *Storage {
	st := Storage{}
	st.db = postgresql.NewPostgresRepository(db)
	return &st
}
