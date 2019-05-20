package postgres

import (
	"database/sql"
	"strings"

	// import the postgres driver
	_ "github.com/lib/pq"
	"github.com/xo/dburl"
)

type Postgres struct {
	conn         *sql.Conn
	db           *sql.DB
	databaseName string
}

func Connect(uri string) (*Postgres, error) {
	db, err := sql.Open("postgres", uri)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	parsed, err := dburl.Parse(uri)
	if err != nil {
		return nil, err
	}

	postgres := Postgres{
		db:           db,
		databaseName: strings.TrimLeft(parsed.Path, "/"),
	}

	return &postgres, nil
}
