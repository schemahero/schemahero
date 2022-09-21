package rqlite

import (
	"net"
	nurl "net/url"

	"github.com/pkg/errors"
	"github.com/rqlite/gorqlite"
)

type RqliteConnection struct {
	db *gorqlite.Connection
}

func Connect(url string) (*RqliteConnection, error) {
	db, err := gorqlite.Open(url)
	if err != nil {
		return nil, err
	}

	rqliteConnection := RqliteConnection{
		db: &db,
	}

	return &rqliteConnection, nil
}

func (s RqliteConnection) Close() error {
	s.db.Close()
	return nil
}

func (m *RqliteConnection) DatabaseName() string {
	return ""
}

func (p *RqliteConnection) EngineVersion() string {
	return ""
}

func UsernameFromURL(url string) (string, error) {
	u, err := nurl.Parse(url)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse url")
	}
	if u.User == nil {
		return "", nil
	}
	return u.User.Username(), nil
}

func PasswordFromURL(url string) (string, error) {
	u, err := nurl.Parse(url)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse url")
	}
	if u.User == nil {
		return "", nil
	}
	pass, _ := u.User.Password()
	return pass, nil
}

func HostnameFromURL(url string) (string, error) {
	u, err := nurl.Parse(url)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse url")
	}
	host, _, err := net.SplitHostPort(u.Host)
	if err != nil {
		return "", errors.Wrap(err, "failed to split host port")
	}
	return host, nil
}

func PortFromURL(url string) (string, error) {
	u, err := nurl.Parse(url)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse url")
	}
	_, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		return "", errors.Wrap(err, "failed to split host port")
	}
	return port, nil
}
