package mysql

import (
	"github.com/pkg/errors"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
)

func PlanMysqlSeedData(uri string, tableName string, seedData *schemasv1alpha4.SeedData) ([]string, error) {
	p, err := Connect(uri)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to mysql")
	}
	defer p.Close()

	return []string{}, errors.New("not implemented")
}
