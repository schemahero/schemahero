package v1alpha4

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDatabase_Postgres_GetConnection_SimpleURI(t *testing.T) {
	d := Database{
		Spec: DatabaseSpec{
			Connection: DatabaseConnection{
				Postgres: &PostgresConnection{
					URI: NewValue("postgres://user:password@host:5432/dbname"),
				},
			},
		},
	}

	ctx := context.Background()

	driver, value, err := d.GetConnection(ctx)

	assert.Equal(t, "postgres", driver)
	assert.Equal(t, "postgres://user:password@host:5432/dbname", value)
	assert.Nil(t, err)
}

func TestDatabase_Postgres_GetConnection_SimpleParams(t *testing.T) {
	d := Database{
		Spec: DatabaseSpec{
			Connection: DatabaseConnection{
				Postgres: &PostgresConnection{
					Host:     NewValue("host"),
					Port:     NewValue("5432"),
					User:     NewValue("user"),
					Password: NewValue("password"),
					DBName:   NewValue("dbname"),
				},
			},
		},
	}

	ctx := context.Background()

	driver, value, err := d.GetConnection(ctx)

	assert.Equal(t, "postgres", driver)
	assert.Equal(t, "postgres://user:password@host:5432/dbname", value)
	assert.Nil(t, err)
}

func TestDatabase_Postgres_GetConnection_SimpleParamsWithSchema(t *testing.T) {
	d := Database{
		Spec: DatabaseSpec{
			Connection: DatabaseConnection{
				Postgres: &PostgresConnection{
					Host:          NewValue("host"),
					Port:          NewValue("5432"),
					User:          NewValue("user"),
					Password:      NewValue("password"),
					DBName:        NewValue("dbname"),
					CurrentSchema: NewValue("public"),
				},
			},
		},
	}

	ctx := context.Background()

	driver, value, err := d.GetConnection(ctx)

	assert.Equal(t, "postgres", driver)
	assert.Equal(t, "postgres://user:password@host:5432/dbname?search_path=public", value)
	assert.Nil(t, err)
}
