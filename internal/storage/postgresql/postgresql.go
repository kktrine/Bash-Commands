package postgresql

import (
	"database/sql"
	_ "github.com/lib/pq"
	"log"
	"time"
)

type Postgres struct {
	Db *sql.DB
}

type Command struct {
	Id      int
	Command string
}

type Outputs struct {
	CommandId int
	Pid       int
	Data      string
	CreatedAt time.Time
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

	createTable := `
	CREATE TABLE IF NOT EXISTS command (
		id SERIAL PRIMARY KEY,
		command TEXT UNIQUE 
	);`
	_, err = db.Exec(createTable)
	if err != nil {
		panic(err)
	}

	return &Postgres{Db: db}
}

func (p *Postgres) InsertCommand(command string) error {
	tx, err := p.Db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec("INSERT INTO command (command) values ($1);", command)

}

func (p *Postgres) Stop() error {
	return p.Db.Close()
}
