package generate

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/schemahero/schemahero/pkg/database/mysql"
	"github.com/schemahero/schemahero/pkg/database/postgres"
)

type Generator struct {
	Driver    string
	URI       string
	DBName    string
	OutputDir string
}

func (g *Generator) RunSync() error {
	fmt.Printf("connecting to %s\n", g.URI)

	if g.Driver == "postgres" {
		pgDb, err := postgres.Connect(g.URI)
		if err != nil {
			return errors.Wrap(err, "failed to connect to postgres")
		}
		return g.generatePostgres(pgDb)
	} else if g.Driver == "mysql" {
		mysqlDb, err := mysql.Connect(g.URI)
		if err != nil {
			return errors.Wrap(err, "failed to connect to mysql")
		}
		return g.generateMysql(mysqlDb)
	} else if g.Driver == "sqlite" {
		// TODO not implemented
		return errors.New("sqlite generate is not implemented")
	} else if g.Driver == "cockroachdb" {
		// TODO not implemented
		return errors.New("cockroachdb generate is not implemented")
	} else if g.Driver == "timescale" {
		// TODO not implemented
		return errors.New("timescale generate is not implemented")
	}

	return fmt.Errorf("unknown database driver for generate: %s", g.Driver)
}

func sanitizeName(name string) string {
	return strings.Replace(name, "_", "-", -1)
}
