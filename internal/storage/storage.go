package storage

import "bash-commands/internal/storage/postgresql"

type Storage struct {
	db *postgresql.Postgres
	//cache *cashe.Cache
}
