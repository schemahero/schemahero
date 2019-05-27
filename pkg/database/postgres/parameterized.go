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

func maybeParseParameterizedColumnType(requestedType string) (string, *int64, error) {
	columnType := ""
	var maxLength *int64

	if strings.HasPrefix(requestedType, "bit varying") {
		columnType = "bit varying"

		r := regexp.MustCompile(`bit varying\s*\((?P<max>\d*)\)`)

		matchGroups := r.FindStringSubmatch(requestedType)
		if len(matchGroups) == 0 {
			max64 := int64(1)
			maxLength = &max64
		} else {
			maxStr := matchGroups[1]
			max, err := strconv.Atoi(maxStr)
			if err != nil {
				return "", maxLength, err
			}
			max64 := int64(max)
			maxLength = &max64
		}
	} else if strings.HasPrefix(requestedType, "bit") {
		columnType = "bit"

		r := regexp.MustCompile(`bit\s*\((?P<max>\d*)\)`)

		matchGroups := r.FindStringSubmatch(requestedType)
		if len(matchGroups) == 0 {
			max64 := int64(1)
			maxLength = &max64
		} else {
			maxStr := matchGroups[1]
			max, err := strconv.Atoi(maxStr)
			if err != nil {
				return "", maxLength, err
			}
			max64 := int64(max)
			maxLength = &max64
		}
	} else if strings.HasPrefix(requestedType, "character varying") {
		columnType = "character varying"

		r := regexp.MustCompile(`character varying\s*\((?P<max>\d*)\)`)

		matchGroups := r.FindStringSubmatch(requestedType)
		if len(matchGroups) == 0 {
			max64 := int64(1)
			maxLength = &max64
		} else {
			maxStr := matchGroups[1]
			max, err := strconv.Atoi(maxStr)
			if err != nil {
				return "", maxLength, err
			}
			max64 := int64(max)
			maxLength = &max64
		}
	} else if strings.HasPrefix(requestedType, "character") {
		columnType = "character"

		r := regexp.MustCompile(`character\s*\((?P<max>\d*)\)`)

		matchGroups := r.FindStringSubmatch(requestedType)
		if len(matchGroups) == 0 {
			max64 := int64(1)
			maxLength = &max64
		} else {
			maxStr := matchGroups[1]
			max, err := strconv.Atoi(maxStr)
			if err != nil {
				return "", maxLength, err
			}
			max64 := int64(max)
			maxLength = &max64
		}
	} else if strings.HasPrefix(requestedType, "timestamp") {
		columnType = "timestamp"

		withPrecisionWithoutTimeZone := regexp.MustCompile(`timestamp\s*\(\s*(?P<precision>.*)\s*\)\s*without time zone`)
		withPrecision := regexp.MustCompile(`timestamp\s*\(\s*(?P<precision>.*)\s*\)`)
		withoutPrecisionWithoutTimeZone := regexp.MustCompile(`timestamp\s*without time zone`)
		withoutPrecision := regexp.MustCompile(`timestamp\s*`)

		withPrecisionMatchGroups := withPrecision.FindStringSubmatch(requestedType)
		withPrecisionWithoutTimeZoneMatchGroups := withPrecisionWithoutTimeZone.FindStringSubmatch(requestedType)
		withoutPrecisionMatchGroups := withoutPrecision.FindStringSubmatch(requestedType)
		withoutPrecisionWithoutTimeZoneMatchGroups := withoutPrecisionWithoutTimeZone.FindStringSubmatch(requestedType)

		if len(withPrecisionWithoutTimeZoneMatchGroups) == 2 {
			columnType = fmt.Sprintf("timestamp (%s) without time zone", withPrecisionWithoutTimeZoneMatchGroups[1])
		} else if len(withoutPrecisionWithoutTimeZoneMatchGroups) == 1 {
			columnType = "timestamp without time zone"
		} else if len(withPrecisionMatchGroups) == 2 {
			columnType = fmt.Sprintf("timestamp (%s)", withPrecisionMatchGroups[1])
		} else if len(withoutPrecisionMatchGroups) == 1 {
			columnType = "timestamp"
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
	}

	return columnType, maxLength, nil
}

func isParameterizedColumnType(requestedType string) bool {
	for _, unparameterizedColumnType := range unparameterizedColumnTypes {
		if unparameterizedColumnType == requestedType {
			return false
		}
	}

	return true
}
