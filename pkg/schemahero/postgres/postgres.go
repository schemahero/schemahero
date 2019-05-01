package postgres

import (
	"database/sql"

	_ "github.com/lib/pq"
)

type Postgres struct {
	conn *sql.Conn
	db   *sql.DB
}

func Connect(uri string) (*Postgres, error) {
	db, err := sql.Open("postgres", uri)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	postgres := Postgres{
		db: db,
	}

	return &postgres, nil
}

func (pg *Postgres) CheckAlive() error {
	return pg.db.Ping()
}
