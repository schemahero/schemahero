package cassandra

import (
	"github.com/pkg/errors"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
)

func PlanCassandraSeedData(hosts []string, username string, password string, keyspace string, seedData *schemasv1alpha4.SeedData) ([]string, error) {
	p, err := Connect(hosts, username, password, keyspace)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to cassandra")
	}
	defer p.Close()

	return []string{}, errors.New("not implemented")
}
