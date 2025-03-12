package v1alpha4

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDatabase_RQLite_GetConnection_SimpleURI(t *testing.T) {
	d := Database{
		Spec: DatabaseSpec{
			Connection: DatabaseConnection{
				RQLite: &RqliteConnection{
					URI: NewValue("user:password@host:5432/dbname"),
				},
			},
		},
	}

	ctx := context.Background()

	driver, value, err := d.GetConnection(ctx)

	assert.Equal(t, "rqlite", driver)
	assert.Equal(t, "user:password@host:5432/dbname", value)
	assert.Nil(t, err)
}

func TestDatabase_RQLite_GetConnection_SimpleParams(t *testing.T) {
	d := Database{
		Spec: DatabaseSpec{
			Connection: DatabaseConnection{
				RQLite: &RqliteConnection{
					Host:     NewValue("host"),
					Port:     NewValue("5432"),
					User:     NewValue("user"),
					Password: NewValue("password"),
				},
			},
		},
	}

	ctx := context.Background()

	driver, value, err := d.GetConnection(ctx)

	assert.Equal(t, "rqlite", driver)
	assert.Equal(t, "https://user:password@host:5432/", value)
	assert.Nil(t, err)
}
