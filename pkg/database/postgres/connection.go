package postgres

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/schemahero/schemahero/pkg/logger"
	"github.com/xo/dburl"
)

type PostgresConnection struct {
	databaseName  string
	engineVersion string
	schema        string   // Default schema to use
	schemas       []string // All schemas to scan

	conn *pgx.Conn
}

func (p *PostgresConnection) DatabaseName() string {
	return p.databaseName
}

func (p *PostgresConnection) EngineVersion() string {
	return p.engineVersion
}

func (p *PostgresConnection) GetConnection() *pgx.Conn {
	return p.conn
}

func Connect(uri string) (*PostgresConnection, error) {
	conn, err := pgx.Connect(context.Background(), uri)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to postgres")
	}

	if err := conn.Ping(context.Background()); err != nil {
		return nil, err
	}

	databaseName, err := DatabaseNameFromURI(uri)
	if err != nil {
		return nil, err
	}

	query := `select version()`
	row := conn.QueryRow(context.Background(), query)
	var reportedVersion string
	if err := row.Scan(&reportedVersion); err != nil {
		return nil, err
	}
	engineVersion, err := parsePostgresVersion(reportedVersion)
	if err != nil {
		logger.Info(err.Error()) // NOTE: this doesnt work with cockroachdb
	}

	schema := "public" // Default to public
	schemas := []string{"public"}
	
	parsed, err := dburl.Parse(uri)
	if err == nil {
		if parsed.Query().Get("schema") != "" {
			schema = parsed.Query().Get("schema")
			schemas = []string{schema}
		} else if parsed.Query().Get("currentSchema") != "" {
			schema = parsed.Query().Get("currentSchema")
			schemas = []string{schema}
		} else if parsed.Query().Get("search_path") != "" {
			searchPath := parsed.Query().Get("search_path")
			pathSchemas := strings.Split(searchPath, ",")
			if len(pathSchemas) > 0 {
				schema = strings.TrimSpace(pathSchemas[0])
				schemas = []string{schema}
			}
		}
		
		if parsed.Query().Get("schemas") != "" {
			schemasList := parsed.Query().Get("schemas")
			schemas = []string{}
			for _, s := range strings.Split(schemasList, ",") {
				schemas = append(schemas, strings.TrimSpace(s))
			}
			if len(schemas) > 0 {
				schema = schemas[0]
			}
		}
	}

	postgresConnection := PostgresConnection{
		databaseName:  databaseName,
		engineVersion: engineVersion,
		schema:        schema,
		schemas:       schemas,
		conn:          conn,
	}

	return &postgresConnection, nil
}

func (p *PostgresConnection) Close() error {
	if p.conn == nil {
		return nil
	}
	return p.conn.Close(context.Background())
}

func DatabaseNameFromURI(uri string) (string, error) {
	parsed, err := dburl.Parse(uri)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse uri")
	}

	return strings.TrimLeft(parsed.Path, "/"), nil
}

func parsePostgresVersion(reportedVersion string) (string, error) {
	//  PostgreSQL 9.5.17 on x86_64-pc-linux-gnu (Debian 9.5.17-1.pgdg90+1), compiled by gcc (Debian 6.3.0-18+deb9u1) 6.3.0 20170516, 64-bit
	r := regexp.MustCompile(`\s*PostgreSQL (?P<major>\d*).(?P<minor>\d*).(?P<patch>\d*)*`)
	matchGroups := r.FindStringSubmatch(reportedVersion)

	if len(matchGroups) == 0 {
		return "", errors.New(`faled to parse postgres version`)
	}

	major := matchGroups[1]
	minor := matchGroups[2]
	patch := matchGroups[3]

	if patch == "" {
		patch = "0"
	}

	return fmt.Sprintf("%s.%s.%s", major, minor, patch), nil
}

func SanitizeArray(idents []string) []string {
	var idents_ []string
	for _, ident := range idents {
		idents_ = append(idents_, pgx.Identifier{ident}.Sanitize())
	}
	return idents_
}
