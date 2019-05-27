package postgres

import (
	"fmt"
	"regexp"
	"strings"
)

func unaliasUnparameterizedColumnType(requestedType string) string {
	switch requestedType {
	case "int8":
		return "bigint"
	case "serial8":
		return "bigserial"
	case "bool":
		return "boolean"
	case "float8":
		return "double precision"
	case "int":
		return "integer"
	case "int4":
		return "integer"
	case "float4":
		return "real"
	case "int2":
		return "smallint"
	case "serial2":
		return "smallserial"
	case "serial4":
		return "serial"
	}

	for _, unparameterizedColumnType := range unparameterizedColumnTypes {
		if unparameterizedColumnType == requestedType {
			return requestedType
		}
	}

	return ""
}

func unaliasParameterizedColumnType(requestedType string) string {
	if strings.HasPrefix(requestedType, "varbit") {
		r := regexp.MustCompile(`varbit\s*\((?P<max>\d*)\)`)

		matchGroups := r.FindStringSubmatch(requestedType)
		if len(matchGroups) == 0 {
			return "bit varying"
		}

		return fmt.Sprintf("bit varying (%s)", matchGroups[1])
	}
	if strings.HasPrefix(requestedType, "char ") || strings.HasPrefix(requestedType, "char(") ||
		requestedType == "character" || requestedType == "char" {
		r := regexp.MustCompile(`char\s*\((?P<len>\d*)\)`)

		matchGroups := r.FindStringSubmatch(requestedType)
		if len(matchGroups) == 0 {
			return "character"
		}

		return fmt.Sprintf("character (%s)", matchGroups[1])
	}
	if strings.HasPrefix(requestedType, "varchar") {
		r := regexp.MustCompile(`varchar\s*\(\s*(?P<max>\d*)\s*\)`)

		matchGroups := r.FindStringSubmatch(requestedType)
		if len(matchGroups) == 0 {
			return "character varying"
		}

		return fmt.Sprintf("character varying (%s)", matchGroups[1])
	}
	if strings.HasPrefix(requestedType, "decimal") {
		precisionAndScale := regexp.MustCompile(`decimal\s*\(\s*(?P<precision>\d*),\s*(?P<scale>\d*)\s*\)`)
		precisionOnly := regexp.MustCompile(`decimal\s*\(\s*(?P<precision>\d*)\s*\)`)

		precisionAndScaleMatchGroups := precisionAndScale.FindStringSubmatch(requestedType)
		precisionOnlyMatchGroups := precisionOnly.FindStringSubmatch(requestedType)

		if len(precisionAndScaleMatchGroups) == 0 && len(precisionOnlyMatchGroups) == 0 {
			return "numeric"
		}

		if len(precisionAndScaleMatchGroups) > 0 {
			return fmt.Sprintf("numeric (%s, %s)", precisionAndScaleMatchGroups[1], precisionAndScaleMatchGroups[2])
		}

		return fmt.Sprintf("numeric (%s)", precisionOnlyMatchGroups[1])
	}
	if strings.HasPrefix(requestedType, "timetz") {
		r := regexp.MustCompile(`timetz\s*\(\s*(?P<precision>.*)\s*\)`)

		matchGroups := r.FindStringSubmatch(requestedType)
		if len(matchGroups) == 0 {
			return "time with time zone"
		}

		return fmt.Sprintf("time (%s) with time zone", matchGroups[1])
	}
	if strings.HasPrefix(requestedType, "timestamptz") {
		r := regexp.MustCompile(`timestamptz\s*\(\s*(?P<precision>.*)\s*\)`)

		matchGroups := r.FindStringSubmatch(requestedType)
		if len(matchGroups) == 0 {
			return "timestamp with time zone"
		}

		return fmt.Sprintf("timestamp (%s) with time zone", matchGroups[1])
	}
	return ""
}
