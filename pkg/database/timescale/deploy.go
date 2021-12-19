package timescale

import (
	"github.com/pkg/errors"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
)

func PlanTimescaleTable(uri string, tableName string, timescaleTableSchema *schemasv1alpha4.TimescaleTableSchema, seedData *schemasv1alpha4.SeedData) ([]string, error) {
	// TODO not implemented
	return nil, errors.New("not implemented")
}
