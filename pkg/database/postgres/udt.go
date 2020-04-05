package postgres

// This function only works for a few built-in types right now
// This needs a better design to make this work past basic arrays
func UDTNameToDataType(udtName string) string {
	switch udtName {
	case "_int4":
		return "integer"
	case "_text":
		return "text"
	}

	return udtName
}
