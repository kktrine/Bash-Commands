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
	Pid       int
	Output    string
}

type CommandRunResult struct {
	Id           int64
	Pid          int
	Output       string
	OutputErrors string
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
	CREATE TABLE IF NOT EXISTS commands (
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
    	created_at TIMESTAMP NOT NULL DEFAULT now(),
    	FOREIGN KEY (command_id) REFERENCES commands(id)
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

func (p Postgres) InsertCommand(command string) (int64, error) {
	tx, err := p.Db.Begin()
	if err != nil {
		return 0, err
	}
	var id int64
	err = tx.QueryRow("INSERT INTO commands (command) values ($1) RETURNING id;", command).Scan(&id)
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

func (p Postgres) SelectOne(id int64) (string, error) {
	res := p.Db.QueryRow("SELECT command FROM commands where id = $1", id)
	err := res.Err()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	var command string
	err = res.Scan(&command)
	if err != nil {
		return "", err
	}
	return command, nil
}

func (p Postgres) SelectAll() (*[]Command, error) {
	rows, err := p.Db.Query("SELECT id, command FROM commands")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var commands []Command
	for rows.Next() {
		var command Command
		err := rows.Scan(&command.Id, &command.Command)
		if err != nil {
			log.Fatal(err)
		}
		commands = append(commands, command)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &commands, nil
}

func (p Postgres) Stop() error {
	return p.Db.Close()
}

func (p Postgres) Delete(id int64) (bool, error) {
	tx, err := p.Db.Begin()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	_, err = tx.Exec("DELETE FROM outputs where command_id = $1", id)
	if err != nil {
		return false, err
	}
	res, err := tx.Exec("DELETE FROM commands where id = $1", id)
	if err != nil {
		return false, err
	}
	err = tx.Commit()
	if err != nil {
		return false, err
	}
	found, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return found > 0, nil
}

func (p Postgres) AddResult(id int64, pid int, stdoutStr string, stderrStr string) error {
	tx, err := p.Db.Begin()
	defer tx.Rollback()
	if err != nil {
		return err
	}
	_, err = tx.Exec("INSERT INTO outputs (command_id, pid, output) values ($1, $2, $3);",
		id, pid, stdoutStr)
	if err != nil {
		return err
	}
	_, err = tx.Exec("INSERT INTO outputs (command_id, pid, output) values ($1, $2, $3);",
		id, pid, stderrStr)
	if err != nil {
		return err
	}
	return nil
}
