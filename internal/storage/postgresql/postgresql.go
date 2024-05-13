package postgresql

import (
	"bash-commands/internal/storage/storageErrors"
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"io"
	"log"
	"os/exec"
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
	println("!!!!!!!!")
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

func (p Postgres) ExecCommand(id int64, command string) (int64, error) {
	tx, err := p.Db.Begin()
	if err != nil {
		return 0, err
	}
	cmd := exec.Command("bash", "-c", command)
	pid := cmd.Process.Pid
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return 0, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return 0, err
	}
	if err := cmd.Start(); err != nil {
		return 0, err
	}

	go func(stdout io.Reader) {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			tx.Exec("INSERT INTO outputs (command_id, pid, output) values ($1, $2, $3);", id, pid, scanner.Text())
		}
	}(stdout)

	go func(stderr io.Reader) {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			tx.Exec("INSERT INTO outputs (command_id, pid, output) values ($1, $2, $3);", id, pid, scanner.Text())
		}
	}(stderr)

	if err := cmd.Wait(); err != nil {
		fmt.Println("Ошибка выполнения команды:", err)
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return int64(pid), nil
}

func (p *Postgres) RunCommandByText(command string) (int64, error) {
	var id int64
	res := p.Db.QueryRow("Select id from commands where command = $1", command)
	err := res.Err()
	if err != nil {
		return 0, err
	}
	err = res.Scan(&id)
	if err != nil {
		return 0, err
	}
	return p.ExecCommand(id, command)
}

func (p *Postgres) RunCommandById(id int64) (int64, error) {
	var command string
	res := p.Db.QueryRow("Select command from commands where id = $1", id)
	err := res.Err()
	if err != nil {
		return 0, err
	}
	err = res.Scan(&command)
	if err != nil {
		return 0, err
	}

	return p.ExecCommand(id, command)
}

func (p *Postgres) Stop() error {
	return p.Db.Close()
}
