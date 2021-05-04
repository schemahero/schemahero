package cassandra

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
)

func CreateTypeStatement(keyspace string, typeName string, typeSchema *schemasv1alpha4.CassandraDataTypeSchema) (string, error) {
	fields := []string{}
	for _, desiredField := range typeSchema.Fields {
		fieldFields, err := cassandraTypeAsInsert(desiredField)
		if err != nil {
			return "", err
		}
		fields = append(fields, fieldFields)
	}

	query := fmt.Sprintf("create type %q (%s)", typeName, strings.Join(fields, ", "))

	return query, nil
}

func CreateTableStatements(keyspace string, tableName string, tableSchema *schemasv1alpha4.CassandraTableSchema) ([]string, error) {
	columns := []string{}
	for _, desiredColumn := range tableSchema.Columns {
		columnsFields, err := cassandraColumnAsInsert(desiredColumn)
		if err != nil {
			return nil, err
		}
		columns = append(columns, columnsFields)
	}

	// primary key
	if tableSchema.PrimaryKey != nil {
		compoundedKeys := []string{}
		for _, primaryKey := range tableSchema.PrimaryKey {
			if len(primaryKey) == 1 {
				compoundedKeys = append(compoundedKeys, primaryKey[0])
				continue
			}

			keyComponent := fmt.Sprintf("(%s)", strings.Join(primaryKey, ", "))
			compoundedKeys = append(compoundedKeys, keyComponent)
		}

		pk := fmt.Sprintf("primary key (%s)", strings.Join(compoundedKeys, ", "))
		columns = append(columns, pk)
	}

	query := fmt.Sprintf(`create table "%s.%s" (%s)`, keyspace, tableName, strings.Join(columns, ", "))

	// clustering
	if tableSchema.ClusteringOrder != nil {
		order := ""
		if tableSchema.ClusteringOrder.IsDescending != nil && *tableSchema.ClusteringOrder.IsDescending {
			order = " desc"
		}
		clustering := fmt.Sprintf("with clustering order by (%s%s)", tableSchema.ClusteringOrder.Column, order)

		query = fmt.Sprintf("%s %s", query, clustering)
	}

	// any specified properties
	tableProperties := []string{}
	if tableSchema.Properties != nil {
		if tableSchema.Properties.BloomFilterFPChance != "" {
			tableProperty := fmt.Sprintf(`bloom_filter_fp_chance = %s`, tableSchema.Properties.BloomFilterFPChance)
			tableProperties = append(tableProperties, tableProperty)
		}
		if tableSchema.Properties.Caching != nil {
			b, err := json.Marshal(tableSchema.Properties.Caching)
			if err != nil {
				return nil, errors.Wrap(err, "failed to marshal caching property")
			}
			tableProperty := fmt.Sprintf(`caching = %s`, b)
			tableProperties = append(tableProperties, tableProperty)
		}
		if tableSchema.Properties.Comment != "" {
			tableProperty := fmt.Sprintf(`comment = '%s'`, tableSchema.Properties.Comment)
			tableProperties = append(tableProperties, tableProperty)
		}
		if tableSchema.Properties.Compaction != nil {
			b, err := json.Marshal(tableSchema.Properties.Compaction)
			if err != nil {
				return nil, errors.Wrap(err, "failed to marshal compaction property")
			}
			tableProperty := fmt.Sprintf(`compaction = %s`, b)
			tableProperties = append(tableProperties, tableProperty)
		}
		if tableSchema.Properties.Compression != nil {
			b, err := json.Marshal(tableSchema.Properties.Compression)
			if err != nil {
				return nil, errors.Wrap(err, "failed to marshal compression property")
			}
			tableProperty := fmt.Sprintf(`compression = %s`, b)
			tableProperties = append(tableProperties, tableProperty)
		}
		if tableSchema.Properties.CRCCheckChance != "" {
			tableProperty := fmt.Sprintf(`crc_check_chance = %s`, tableSchema.Properties.CRCCheckChance)
			tableProperties = append(tableProperties, tableProperty)
		}
		if tableSchema.Properties.DCLocalReadRepairChance != "" {
			tableProperty := fmt.Sprintf(`dclocal_read_repair_chance = %s`, tableSchema.Properties.DCLocalReadRepairChance)
			tableProperties = append(tableProperties, tableProperty)
		}
		if tableSchema.Properties.DefaultTTL != nil {
			tableProperty := fmt.Sprintf(`default_time_to_live = %d`, *tableSchema.Properties.DefaultTTL)
			tableProperties = append(tableProperties, tableProperty)
		}
		if tableSchema.Properties.GCGraceSeconds != nil {
			tableProperty := fmt.Sprintf(`grace_seconds = %d`, *tableSchema.Properties.GCGraceSeconds)
			tableProperties = append(tableProperties, tableProperty)
		}
		if tableSchema.Properties.MaxIndexInterval != nil {
			tableProperty := fmt.Sprintf(`max_index_interval = %d`, *tableSchema.Properties.MaxIndexInterval)
			tableProperties = append(tableProperties, tableProperty)
		}
		if tableSchema.Properties.MemtableFlushPeriodMS != nil {
			tableProperty := fmt.Sprintf(`memtable_flush_period_in_ms = %d`, *tableSchema.Properties.MemtableFlushPeriodMS)
			tableProperties = append(tableProperties, tableProperty)
		}
		if tableSchema.Properties.MinIndexInterval != nil {
			tableProperty := fmt.Sprintf(`min_index_interval = %d`, *tableSchema.Properties.MinIndexInterval)
			tableProperties = append(tableProperties, tableProperty)
		}
		if tableSchema.Properties.ReadRepairChance != "" {
			tableProperty := fmt.Sprintf(`read_repair_chance = %s`, tableSchema.Properties.ReadRepairChance)
			tableProperties = append(tableProperties, tableProperty)
		}
		if tableSchema.Properties.SpeculativeRetry != "" {
			tableProperty := fmt.Sprintf(`speculative_retry = '%s'`, tableSchema.Properties.SpeculativeRetry)
			tableProperties = append(tableProperties, tableProperty)
		}
	}

	if len(tableProperties) > 0 {
		query = fmt.Sprintf("%s with %s", query, strings.Join(tableProperties, " AND "))
	}

	return []string{query}, nil
}
