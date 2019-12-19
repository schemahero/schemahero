package v1alpha3

import (
	"github.com/schemahero/schemahero/pkg/apis/databases/v1alpha2"
)

func ConvertFromV1Alpha2(instance *v1alpha2.Database) *Database {
	converted := &Database{
		TypeMeta:   instance.TypeMeta,
		ObjectMeta: instance.ObjectMeta,
	}

	converted.Status = DatabaseStatus{
		IsConnected: instance.Status.IsConnected,
		LastPing:    instance.Status.LastPing,
	}

	connection := DatabaseConnection{}
	if instance.Connection.Postgres != nil {
		postgres := PostgresConnection{
			URI: ConvertValueOrValueFromFromV1Alpha2(instance.Connection.Postgres.URI),
		}
		connection.Postgres = &postgres
	}

	converted.Connection = connection

	if instance.SchemaHero != nil {
		schemaHero := SchemaHero(*instance.SchemaHero)
		converted.SchemaHero = &schemaHero
	}

	return converted
}
