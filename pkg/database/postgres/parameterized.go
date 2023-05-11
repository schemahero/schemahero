package postgres

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var unparameterizedColumnTypes = []string{
	"bigint",
	"bigserial",
	"boolean",
	"box",
	"bytea",
	"cidr",
	"circle",
	"citext",
	"date",
	"double precision",
	"inet",
	"integer",
	"json",
	"jsonb",
	"line",
	"lseg",
	"macaddr",
	"money",
	"path",
	"pg_lsn",
	"point",
	"polygon",
	"real",
	"smallint",
	"smallserial",
	"serial",
	"text",
	"tsquery",
	"tsvector",
	"txid_snapshot",
	"uuid",
	"xml",
}

func maybeParseParameterizedColumnType(requestedType string) (string, error) {
	columnType := ""

	if strings.HasPrefix(requestedType, "bit varying") {
		columnType = "bit varying"

		r := regexp.MustCompile(`bit varying\s*\((?P<max>\d*)\)`)

		matchGroups := r.FindStringSubmatch(requestedType)
		if len(matchGroups) == 0 {
			columnType = "bit varying (1)"
		} else {
			maxStr := matchGroups[1]
			max, err := strconv.Atoi(maxStr)
			if err != nil {
				return "", err
			}
			columnType = fmt.Sprintf("bit varying (%d)", max)
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
	} else if strings.HasPrefix(requestedType, "character varying") {
		columnType = "character varying"

		r := regexp.MustCompile(`character varying\s*\((?P<max>\d*)\)`)

		matchGroups := r.FindStringSubmatch(requestedType)
		if len(matchGroups) == 0 {
			columnType = "character varying (1)"
		} else {
			maxStr := matchGroups[1]
			max, err := strconv.Atoi(maxStr)
			if err != nil {
				return "", err
			}
			columnType = fmt.Sprintf("character varying (%d)", max)
		}
	} else if strings.HasPrefix(requestedType, "character") {
		columnType = "character"

		r := regexp.MustCompile(`character\s*\((?P<max>\d*)\)`)

		matchGroups := r.FindStringSubmatch(requestedType)
		if len(matchGroups) == 0 {
			columnType = "character (1)"
		} else {
			maxStr := matchGroups[1]
			max, err := strconv.Atoi(maxStr)
			if err != nil {
				return "", err
			}
			columnType = fmt.Sprintf("character (%d)", max)
		}
	} else if strings.HasPrefix(requestedType, "timestamp") {
		columnType = "timestamp"

		withPrecisionWithoutTimeZone := regexp.MustCompile(`timestamp\s*\(\s*(?P<precision>.*)\s*\)\s*without time zone`)
		withPrecision := regexp.MustCompile(`timestamp\s*\(\s*(?P<precision>.*)\s*\)`)
		withoutPrecisionWithoutTimeZone := regexp.MustCompile(`timestamp\s*without time zone`)
		withPrecisionWithTimeZone := regexp.MustCompile(`timestamp\s*\(\s*(?P<precision>.*)\s*\)\s*with time zone`)
		withoutPrecisionWithTimeZone := regexp.MustCompile(`timestamp\s*with time zone`)
		withoutPrecision := regexp.MustCompile(`timestamp\s*`)

		withPrecisionMatchGroups := withPrecision.FindStringSubmatch(requestedType)
		withPrecisionWithoutTimeZoneMatchGroups := withPrecisionWithoutTimeZone.FindStringSubmatch(requestedType)
		withPrecisionWithTimeZoneMatchGroups := withPrecisionWithTimeZone.FindStringSubmatch(requestedType)
		withoutPrecisionMatchGroups := withoutPrecision.FindStringSubmatch(requestedType)
		withoutPrecisionWithoutTimeZoneMatchGroups := withoutPrecisionWithoutTimeZone.FindStringSubmatch(requestedType)
		withoutPrecisionWithTimeZoneMatchGroups := withoutPrecisionWithTimeZone.FindStringSubmatch(requestedType)

		if len(withPrecisionWithoutTimeZoneMatchGroups) == 2 {
			columnType = fmt.Sprintf("timestamp (%s) without time zone", withPrecisionWithoutTimeZoneMatchGroups[1])
		} else if len(withoutPrecisionWithoutTimeZoneMatchGroups) == 1 {
			columnType = "timestamp without time zone"
		} else if len(withPrecisionWithTimeZoneMatchGroups) == 2 {
			columnType = fmt.Sprintf("timestamp (%s) with time zone", withPrecisionWithTimeZoneMatchGroups[1])
		} else if len(withoutPrecisionWithTimeZoneMatchGroups) == 1 {
			columnType = "timestamp with time zone"
		} else if len(withPrecisionMatchGroups) == 2 {
			columnType = fmt.Sprintf("timestamp (%s)", withPrecisionMatchGroups[1])
		} else if len(withoutPrecisionMatchGroups) == 1 {
			columnType = "timestamp"
		}
	} else if strings.HasPrefix(requestedType, "time") {
		columnType = "time"

		withPrecisionWithoutTimeZone := regexp.MustCompile(`time\s*\(\s*(?P<precision>.*)\s*\)\s*without time zone`)
		withPrecision := regexp.MustCompile(`time\s*\(\s*(?P<precision>.*)\s*\)`)
		withoutPrecisionWithoutTimeZone := regexp.MustCompile(`time\s*without time zone`)
		withPrecisionWithTimeZone := regexp.MustCompile(`timestamp\s*\(\s*(?P<precision>.*)\s*\)\s*with time zone`)
		withoutPrecision := regexp.MustCompile(`time\s*`)
		withoutPrecisionWithTimeZone := regexp.MustCompile(`time\s*with time zone`)

		withPrecisionMatchGroups := withPrecision.FindStringSubmatch(requestedType)
		withPrecisionWithoutTimeZoneMatchGroups := withPrecisionWithoutTimeZone.FindStringSubmatch(requestedType)
		withPrecisionWithTimeZoneMatchGroups := withPrecisionWithTimeZone.FindStringSubmatch(requestedType)
		withoutPrecisionMatchGroups := withoutPrecision.FindStringSubmatch(requestedType)
		withoutPrecisionWithoutTimeZoneMatchGroups := withoutPrecisionWithoutTimeZone.FindStringSubmatch(requestedType)
		withoutPrecisionWithTimeZoneMatchGroups := withoutPrecisionWithTimeZone.FindStringSubmatch(requestedType)

		if len(withPrecisionWithoutTimeZoneMatchGroups) == 2 {
			columnType = fmt.Sprintf("time (%s) without time zone", withPrecisionWithoutTimeZoneMatchGroups[1])
		} else if len(withoutPrecisionWithoutTimeZoneMatchGroups) == 1 {
			columnType = "time without time zone"
		} else if len(withPrecisionWithTimeZoneMatchGroups) == 2 {
			columnType = fmt.Sprintf("time (%s) with time zone", withPrecisionWithoutTimeZoneMatchGroups[1])
		} else if len(withoutPrecisionWithTimeZoneMatchGroups) == 1 {
			columnType = "time with time zone"
		} else if len(withPrecisionMatchGroups) == 2 {
			columnType = fmt.Sprintf("time (%s)", withPrecisionMatchGroups[1])
		} else if len(withoutPrecisionMatchGroups) == 1 {
			columnType = "time"
		}
	} else if strings.HasPrefix(requestedType, "numeric") {
		columnType = "numeric"

		precisionAndScale := regexp.MustCompile(`numeric\s*\(\s*(?P<precision>\d*),\s*(?P<scale>\d*)\s*\)`)
		precisionOnly := regexp.MustCompile(`numeric\s*\(\s*(?P<precision>\d*)\s*\)`)

		precisionAndScaleMatchGroups := precisionAndScale.FindStringSubmatch(requestedType)
		precisionOnlyMatchGroups := precisionOnly.FindStringSubmatch(requestedType)

		if len(precisionAndScaleMatchGroups) == 3 {
			columnType = fmt.Sprintf("numeric (%s, %s)", precisionAndScaleMatchGroups[1], precisionAndScaleMatchGroups[2])
		} else if len(precisionOnlyMatchGroups) == 2 {
			columnType = fmt.Sprintf("numeric (%s)", precisionOnlyMatchGroups[1])
		}
	} else if strings.HasPrefix(requestedType, "vector") {
		columnType = "vector"

		size := regexp.MustCompile(`vector\s*\(\s*(?P<size>\d*)\s*\)`)
		sizeMatchGroups := size.FindStringSubmatch(requestedType)

		if len(sizeMatchGroups) == 2 {
			columnType = fmt.Sprintf("vector (%s)", sizeMatchGroups[1])
		}
	}

	return columnType, nil
}

func isParameterizedColumnType(requestedType string) bool {
	for _, unparameterizedColumnType := range unparameterizedColumnTypes {
		if unparameterizedColumnType == requestedType {
			return false
		}
	}

	return true
}
