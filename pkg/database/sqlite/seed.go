package sqlite

import (
	"github.com/pkg/errors"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
)

func PlanSqliteSeedData(uri string, tableName string, seedData *schemasv1alpha4.SeedData) ([]string, error) {
	p, err := Connect(uri)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to sqlite")
	}
	defer p.Close()

	return []string{}, errors.New("not implemented")
}
