package cassandra

func unaliasUnparameterizedColumnType(requestedType string) string {
	switch requestedType {
	case "varchar":
		return "text"
	}

	// TODO

	// for _, unparameterizedColumnType := range unparameterizedColumnTypes {
	// 	if unparameterizedColumnType == requestedType {
	// 		return requestedType
	// 	}
	// }

	return requestedType
}
