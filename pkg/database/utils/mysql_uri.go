package utils

import (
	"strings"

	mysqldriver "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
)

// MySQL URI parsing utilities
// These are kept in the main tree for use by CLI commands

func MySQLDatabaseNameFromURI(uri string) (string, error) {
	cfg, err := mysqldriver.ParseDSN(uri)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse database uri")
	}

	return cfg.DBName, nil
}

func MySQLUsernameFromURI(uri string) (string, error) {
	cfg, err := mysqldriver.ParseDSN(uri)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse database uri")
	}

	return cfg.User, nil
}

func MySQLPasswordFromURI(uri string) (string, error) {
	cfg, err := mysqldriver.ParseDSN(uri)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse database uri")
	}

	return cfg.Passwd, nil
}

func MySQLHostnameFromURI(uri string) (string, error) {
	cfg, err := mysqldriver.ParseDSN(uri)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse database uri")
	}

	// TODO this is very happy path
	addrAndPort := strings.Split(cfg.Addr, ":")
	return addrAndPort[0], nil
}

func MySQLPortFromURI(uri string) (string, error) {
	cfg, err := mysqldriver.ParseDSN(uri)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse database uri")
	}

	// TODO this is very happy path
	addrAndPort := strings.Split(cfg.Addr, ":")

	if len(addrAndPort) > 1 {
		return addrAndPort[1], nil
	}
	return "3306", nil
}
