package v1alpha4

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDatabase_Cassandra_GetConnection_SimpleParams(t *testing.T) {
	d := Database{
		Spec: DatabaseSpec{
			Connection: DatabaseConnection{
				Cassandra: &CassandraConnection{
					Hosts:    []string{"host1", "host2"},
					Keyspace: NewValue("keyspace"),
					Username: NewValue("username"),
					Password: NewValue("password"),
				},
			},
		},
	}

	ctx := context.Background()

	driver, value, err := d.GetConnection(ctx)

	// populate the strings and change NotNil to Nil once implemented
	assert.Equal(t, "", driver)
	assert.Equal(t, "", value)
	assert.NotNil(t, err)
}
