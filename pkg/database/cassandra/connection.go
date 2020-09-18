package cassandra

import (
	"github.com/gocql/gocql"
	"github.com/pkg/errors"
)

type CassandraConnection struct {
	session  *gocql.Session
	keyspace string
}

func Connect(hosts []string, username string, password string, keyspace string) (*CassandraConnection, error) {
	cluster := gocql.NewCluster(hosts...)

	if username != "" {
		cluster.Authenticator = gocql.PasswordAuthenticator{
			Username: username,
			Password: password,
		}
	}

	session, err := cluster.CreateSession()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cassandra session")
	}

	cassandraConnection := CassandraConnection{
		session:  session,
		keyspace: keyspace,
	}

	return &cassandraConnection, nil
}

func (c *CassandraConnection) Close() {
	c.session.Close()
}
