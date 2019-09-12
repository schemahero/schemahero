package postgres

import (
	"database/sql"
	goerrors "errors"
	"fmt"
	"regexp"
	"strings"

	// import the postgres driver
	_ "github.com/lib/pq"
	"github.com/xo/dburl"
)

type PostgresConnection struct {
	conn          *sql.Conn
	db            *sql.DB
	databaseName  string
	engineVersion string
}

func (p *PostgresConnection) GetConnection() *sql.Conn {
	return p.conn
}

func (p *PostgresConnection) GetDB() *sql.DB {
	return p.db
}

func (p *PostgresConnection) DatabaseName() string {
	return p.databaseName
}

func (p *PostgresConnection) EngineVersion() string {
	return p.engineVersion
}

func Connect(uri string) (*PostgresConnection, error) {
	db, err := sql.Open("postgres", uri)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	databaseName, err := DatabaseNameFromURI(uri)
	if err != nil {
		return nil, err
	}

	query := `select version()`
	row := db.QueryRow(query)
	var reportedVersion string
	if err := row.Scan(&reportedVersion); err != nil {
		return nil, err
	}
	engineVersion, err := parsePostgresVersion(reportedVersion)
	if err != nil {
		log.Info(err.Error()) // NOTE: this doesnt work with cockroachdb
	}

	postgresConnection := PostgresConnection{
		db:            db,
		databaseName:  databaseName,
		engineVersion: engineVersion,
	}

	return &postgresConnection, nil
}

func DatabaseNameFromURI(uri string) (string, error) {
	parsed, err := dburl.Parse(uri)
	if err != nil {
		return "", err
	}

	return strings.TrimLeft(parsed.Path, "/"), nil
}

func parsePostgresVersion(reportedVersion string) (string, error) {
	//  PostgreSQL 9.5.17 on x86_64-pc-linux-gnu (Debian 9.5.17-1.pgdg90+1), compiled by gcc (Debian 6.3.0-18+deb9u1) 6.3.0 20170516, 64-bit
	r := regexp.MustCompile(`\s*PostgreSQL (?P<major>\d*).(?P<minor>\d*).(?P<patch>\d*)*`)
	matchGroups := r.FindStringSubmatch(reportedVersion)

	if len(matchGroups) == 0 {
		return "", goerrors.New(`faled to parse postgres version`)
	}

	major := matchGroups[1]
	minor := matchGroups[2]
	patch := matchGroups[3]

	if patch == "" {
		patch = "0"
	}

	return fmt.Sprintf("%s.%s.%s", major, minor, patch), nil
}
