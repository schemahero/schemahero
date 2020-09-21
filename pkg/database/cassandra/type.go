package cassandra

import (
	"fmt"

	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
)

func cassandraTypeAsInsert(field *schemasv1alpha4.CassandraField) (string, error) {
	// TODO before merge!  find the right gocql sanitize methods to call

	result := fmt.Sprintf("%s %s", field.Name, field.Type)

	return result, nil
}
