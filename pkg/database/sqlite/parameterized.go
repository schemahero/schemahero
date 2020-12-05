package sqlite

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var unparameterizedColumnTypes = []string{
	"date",
	"datetime",
	"timestamp",
	"tinyblob",
	"tinytext",
	"mediumblob",
	"mediumtext",
	"longblob",
	"longtext",
	"blob",
	"text",
}

func isParameterizedColumnType(requestedType string) bool {
	for _, unparameterizedColumnType := range unparameterizedColumnTypes {
		if unparameterizedColumnType == requestedType {
			return false
		}
	}

	return true
}

func maybeParseParameterizedColumnType(requestedType string) (string, error) {
	columnType := ""

	if strings.HasPrefix(requestedType, "varchar") {
		columnType = "varchar"

		r := regexp.MustCompile(`varchar\s*\((?P<max>\d*)\)`)

		matchGroups := r.FindStringSubmatch(requestedType)
		if len(matchGroups) == 0 {
			columnType = "varchar (1)"
		} else {
			maxStr := matchGroups[1]
			max, err := strconv.Atoi(maxStr)
			if err != nil {
				return "", err
			}
			columnType = fmt.Sprintf("varchar (%d)", max)
		}
	} else if strings.HasPrefix(requestedType, "char") {
		columnType = "char"

		if strings.Contains(requestedType, "character") {
			requestedType = strings.Replace(requestedType, "character", "char", -1)
		}

		r := regexp.MustCompile(`char\s*\((?P<max>\d*)\)`)

		matchGroups := r.FindStringSubmatch(requestedType)

		if len(matchGroups) == 0 {
			columnType = "char (11)"
		} else {
			maxStr := matchGroups[1]
			max, err := strconv.Atoi(maxStr)
			if err != nil {
				return "", err
			}
			columnType = fmt.Sprintf("char (%d)", max)
		}
	} else if strings.HasPrefix(requestedType, "tinyint") {
		columnType = "tinyint"

		r := regexp.MustCompile(`tinyint\s*\((?P<max>\d*)\)`)

		matchGroups := r.FindStringSubmatch(requestedType)
		if len(matchGroups) == 0 {
			columnType = "tinyint (1)"
		} else {
			maxStr := matchGroups[1]
			max, err := strconv.Atoi(maxStr)
			if err != nil {
				return "", err
			}
			columnType = fmt.Sprintf("tinyint (%d)", max)
		}
	} else if strings.HasPrefix(requestedType, "bit") {
		columnType = "bit"

		r := regexp.MustCompile(`bit\s*\((?P<max>\d*)\)`)

		matchGroups := r.FindStringSubmatch(requestedType)
		if len(matchGroups) == 0 {
			columnType = "bit (1)"
		} else {
			maxStr := matchGroups[1]
			max, err := strconv.Atoi(maxStr)
			if err != nil {
				return "", err
			}
			columnType = fmt.Sprintf("bit (%d)", max)
		}
	} else if strings.HasPrefix(requestedType, "smallint") {
		columnType = "smallint"

		r := regexp.MustCompile(`smallint\s*\((?P<max>\d*)\)`)

		matchGroups := r.FindStringSubmatch(requestedType)
		if len(matchGroups) == 0 {
			columnType = "smallint (5)"
		} else {
			maxStr := matchGroups[1]
			max, err := strconv.Atoi(maxStr)
			if err != nil {
				return "", err
			}
			columnType = fmt.Sprintf("smallint (%d)", max)
		}
	} else if strings.HasPrefix(requestedType, "mediumint") {
		columnType = "mediumint"

		r := regexp.MustCompile(`mediumint\s*\((?P<max>\d*)\)`)

		matchGroups := r.FindStringSubmatch(requestedType)
		if len(matchGroups) == 0 {
			columnType = "mediumint (9)"
		} else {
			maxStr := matchGroups[1]
			max, err := strconv.Atoi(maxStr)
			if err != nil {
				return "", err
			}
			columnType = fmt.Sprintf("mediumint (%d)", max)
		}
	} else if strings.HasPrefix(requestedType, "int") {
		columnType = "int"

		r := regexp.MustCompile(`int\s*\((?P<max>\d*)\)`)

		matchGroups := r.FindStringSubmatch(requestedType)
		if len(matchGroups) == 0 {
			columnType = "int (11)"
		} else {
			maxStr := matchGroups[1]
			max, err := strconv.Atoi(maxStr)
			if err != nil {
				return "", err
			}
			columnType = fmt.Sprintf("int (%d)", max)
		}
	} else if strings.HasPrefix(requestedType, "bigint") {
		columnType = "bigint"

		r := regexp.MustCompile(`bigint\s*\((?P<max>\d*)\)`)

		matchGroups := r.FindStringSubmatch(requestedType)
		if len(matchGroups) == 0 {
			columnType = "bigint (20)"
		} else {
			maxStr := matchGroups[1]
			max, err := strconv.Atoi(maxStr)
			if err != nil {
				return "", err
			}
			columnType = fmt.Sprintf("bigint (%d)", max)
		}
	} else if strings.HasPrefix(requestedType, "decimal") {
		columnType = "decimal"

		precisionAndScale := regexp.MustCompile(`decimal\s*\(\s*(?P<precision>\d*),\s*(?P<scale>\d*)\s*\)`)
		precisionOnly := regexp.MustCompile(`decimal\s*\(\s*(?P<precision>\d*)\s*\)`)

		precisionAndScaleMatchGroups := precisionAndScale.FindStringSubmatch(requestedType)
		precisionOnlyMatchGroups := precisionOnly.FindStringSubmatch(requestedType)

		if len(precisionAndScaleMatchGroups) == 3 {
			columnType = fmt.Sprintf("decimal (%s, %s)", precisionAndScaleMatchGroups[1], precisionAndScaleMatchGroups[2])
		} else if len(precisionOnlyMatchGroups) == 2 {
			columnType = fmt.Sprintf("decimal (%s, 0)", precisionOnlyMatchGroups[1])
		} else {
			columnType = "decimal (10, 0)"
		}
	} else if strings.HasPrefix(requestedType, "float") {
		columnType = "float"

		precisionAndScale := regexp.MustCompile(`float\s*\(\s*(?P<precision>\d*),\s*(?P<scale>\d*)\s*\)`)
		precisionOnly := regexp.MustCompile(`float\s*\(\s*(?P<precision>\d*)\s*\)`)

		precisionAndScaleMatchGroups := precisionAndScale.FindStringSubmatch(requestedType)
		precisionOnlyMatchGroups := precisionOnly.FindStringSubmatch(requestedType)

		if len(precisionAndScaleMatchGroups) == 3 {
			columnType = fmt.Sprintf("float (%s, %s)", precisionAndScaleMatchGroups[1], precisionAndScaleMatchGroups[2])
		} else if len(precisionOnlyMatchGroups) == 2 {
			precision, err := strconv.Atoi(precisionOnlyMatchGroups[1])
			if err != nil {
				return "", err
			}

			if precision > 24 {
				columnType = "double"
			}
		} else {
			columnType = "float"
		}
	} else if strings.HasPrefix(requestedType, "double") {
		columnType = "double"

		precisionAndScale := regexp.MustCompile(`double\s*\(\s*(?P<precision>\d*),\s*(?P<scale>\d*)\s*\)`)
		precisionAndScaleMatchGroups := precisionAndScale.FindStringSubmatch(requestedType)

		if len(precisionAndScaleMatchGroups) == 3 {
			columnType = fmt.Sprintf("double (%s, %s)", precisionAndScaleMatchGroups[1], precisionAndScaleMatchGroups[2])
		} else {
			columnType = "double"
		}
	} else if strings.HasPrefix(requestedType, "binary") {
		columnType = "binary"

		r := regexp.MustCompile(`binary\s*\((?P<max>\d*)\)`)

		matchGroups := r.FindStringSubmatch(requestedType)
		if len(matchGroups) == 0 {
			columnType = "binary (1)"
		} else {
			maxStr := matchGroups[1]
			max, err := strconv.Atoi(maxStr)
			if err != nil {
				return "", err
			}
			columnType = fmt.Sprintf("binary (%d)", max)
		}
	} else if strings.HasPrefix(requestedType, "varbinary") {
		columnType = "varbinary"

		r := regexp.MustCompile(`varbinary\s*\((?P<max>\d*)\)`)

		matchGroups := r.FindStringSubmatch(requestedType)
		if len(matchGroups) == 0 {
			columnType = "varbinary (1)"
		} else {
			maxStr := matchGroups[1]
			max, err := strconv.Atoi(maxStr)
			if err != nil {
				return "", err
			}
			columnType = fmt.Sprintf("varbinary (%d)", max)
		}
	} else if strings.HasPrefix(requestedType, "blob") {
		columnType = "blob"

		r := regexp.MustCompile(`blob\s*\((?P<max>\d*)\)`)

		matchGroups := r.FindStringSubmatch(requestedType)
		if len(matchGroups) > 0 {
			maxStr := matchGroups[1]
			max, err := strconv.Atoi(maxStr)
			if err != nil {
				return "", err
			}
			columnType = fmt.Sprintf("blob (%d)", max)
		}
	} else if strings.HasPrefix(requestedType, "text") {
		columnType = "text"

		r := regexp.MustCompile(`text\s*\((?P<max>\d*)\)`)

		matchGroups := r.FindStringSubmatch(requestedType)
		if len(matchGroups) > 0 {
			maxStr := matchGroups[1]
			max, err := strconv.Atoi(maxStr)
			if err != nil {
				return "", err
			}
			columnType = fmt.Sprintf("text (%d)", max)
		}
	}

	return columnType, nil
}
