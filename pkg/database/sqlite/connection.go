package sqlite

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type SqliteConnection struct {
	db *sql.DB
}

func Connect(dsn string) (*SqliteConnection, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}

	sqliteConnection := SqliteConnection{
		db: db,
	}

	return &sqliteConnection, nil
}

func (s SqliteConnection) Close() {
	s.db.Close()
}
