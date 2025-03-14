package v1alpha4

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDatabase_Mysql_GetConnection_SimpleURI(t *testing.T) {
	d := Database{
		Spec: DatabaseSpec{
			Connection: DatabaseConnection{
				Mysql: &MysqlConnection{
					URI: NewValue("user:password@host:5432/dbname"),
				},
			},
		},
	}

	ctx := context.Background()

	driver, value, err := d.GetConnection(ctx)

	assert.Equal(t, "mysql", driver)
	assert.Equal(t, "user:password@host:5432/dbname", value)
	assert.Nil(t, err)
}

func TestDatabase_Mysql_GetConnection_SimpleParams(t *testing.T) {
	d := Database{
		Spec: DatabaseSpec{
			Connection: DatabaseConnection{
				Mysql: &MysqlConnection{
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

	assert.Equal(t, "mysql", driver)
	assert.Equal(t, "user:password@tcp(host:5432)/dbname", value)
	assert.Nil(t, err)
}
