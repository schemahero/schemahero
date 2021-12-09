package v1alpha4

type SeedDataValue struct {
	Int *int    `json:"int" yaml:"int"`
	Str *string `json:"str" yaml:"str"`
}

type Column struct {
	Column string        `json:"column" yaml:"column"`
	Value  SeedDataValue `json:"value" yaml:"value"`
}

type SeedDataRow struct {
	Columns []Column `json:"columns" yaml:"columns"`
}

type SeedData struct {
	Rows []SeedDataRow `json:"rows" yaml:"rows"`
}
