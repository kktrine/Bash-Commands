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

func (p Postgres) ExecCommand(id int64, command string) (*CommandRunResult, error) {
	tx, err := p.Db.Begin()
	if err != nil {
		return nil, err
	}
	cmd := exec.Command("bash", "-c", command)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	pid := cmd.Process.Pid
	var stdoutStr, stderrStr string
	go func(stdout io.Reader) {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			stdoutStr += line + "\n"
			tx.Exec("INSERT INTO outputs (command_id, pid, output) values ($1, $2, $3);",
				id, pid, line)
		}
	}(stdout)

	go func(stderr io.Reader) {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			stderrStr += line + "\n"
			tx.Exec("INSERT INTO outputs (command_id, pid, output) values ($1, $2, $3);",
				id, pid, line)
		}
	}(stderr)

	if err := cmd.Wait(); err != nil {
		fmt.Println("Ошибка выполнения команды:", err)
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &CommandRunResult{
		Id:           id,
		Pid:          pid,
		Output:       stdoutStr,
		OutputErrors: stderrStr,
	}, nil
}

func (p Postgres) RunCommandById(id int64) (*CommandRunResult, error) {
	var command string
	res := p.Db.QueryRow("Select command from commands where id = $1", id)
	err := res.Err()
	if err != nil {
		return nil, err
	}
	err = res.Scan(&command)
	if err != nil {
		return nil, err
	}

	return p.ExecCommand(id, command)
}

func (p *Postgres) Stop() error {
	return p.Db.Close()
}
