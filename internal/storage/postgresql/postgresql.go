package postgresql

import (
	"bash-commands/internal/storage/storageErrors"
	"database/sql"
	"errors"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"log"
)

type Postgres struct {
	Db *sql.DB
}

type Command struct {
	Id      int64
	Command string
}

type Outputs struct {
	id        int64
	CommandId int64
	Pid       int64
	Output    string
}

func NewPostgresRepository(bdAttributes string) *Postgres {
	db, err := sql.Open("postgres", bdAttributes)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		return nil
	}
	createTableCommands := `
	CREATE TABLE IF NOT EXISTS command (
		id SERIAL PRIMARY KEY,
		command TEXT UNIQUE NOT NULL CHECK (command <> '')
	);`
	_, err = tx.Exec(createTableCommands)
	if err != nil {
		_ = tx.Rollback()
		panic(err)
	}
	createTableOutputs := `
	CREATE TABLE IF NOT EXISTS outputs (
    	id SERIAL PRIMARY KEY,
    	command_id INT,
    	pid INT,
    	output TEXT,
    	FOREIGN KEY (command_id) REFERENCES command(id)
	);`
	_, err = tx.Exec(createTableOutputs)
	if err != nil {
		_ = tx.Rollback()
		panic(err)
	}
	err = tx.Commit()
	if err != nil {
		panic(err)
	}
	return &Postgres{Db: db}
}

func (p *Postgres) InsertCommand(command string) (int64, error) {
	tx, err := p.Db.Begin()
	if err != nil {
		return 0, err
	}
	var id int64
	err = tx.QueryRow("INSERT INTO command (command) values ($1) RETURNING id;", command).Scan(&id)
	if err != nil {
		errRollB := tx.Rollback()
		if errRollB != nil {
			return 0, err
		}
		var pgErr *pq.Error
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return 0, storageErrors.ErrDuplicateEntry
			}
		}
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return id, nil
}

func (p *Postgres) Stop() error {
	return p.Db.Close()
}
