package cassandra

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/schemahero/schemahero/pkg/database/types"
)

func PlanCassandraType(hosts []string, username string, password string, keyspace string, typeName string, cassandraTypeSchema *schemasv1alpha4.CassandraDataTypeSchema) ([]string, error) {
	c, err := Connect(hosts, username, password, keyspace)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to cassandra")
	}
	defer c.Close()

	// determine if the type exists
	query := `select count(1) from system_schema.types where keyspace_name=? and type_name=?`
	row := c.session.Query(query, keyspace, typeName)
	typeExists := 0
	if err := row.Scan(&typeExists); err != nil {
		return nil, errors.Wrap(err, "failed to scan")
	}

	if typeExists == 0 && cassandraTypeSchema.IsDeleted {
		return []string{}, nil
	} else if typeExists > 0 && cassandraTypeSchema.IsDeleted {
		return []string{
			fmt.Sprintf(`drop ttype %s.%s`, keyspace, typeName),
		}, nil
	}

	if typeExists == 0 {
		// shortcut to just create it
		query, err := CreateTypeStatement(keyspace, typeName, cassandraTypeSchema)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create type statement")
		}

		return []string{query}, nil
	}

	return nil, errors.New("not implemented")
}

func PlanCassandraView(hosts []string, username string, password string, keyspace string, viewName string, cassandraViewSchema *schemasv1alpha4.NotImplementedViewSchema) ([]string, error) {
	return nil, errors.New("not implemented")
}

func PlanCassandraTable(hosts []string, username string, password string, keyspace string, tableName string, cassandraTableSchema *schemasv1alpha4.CassandraTableSchema, seedData *schemasv1alpha4.SeedData) ([]string, error) {
	c, err := Connect(hosts, username, password, keyspace)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to cassandra")
	}
	defer c.Close()

	// determine if the table exists
	query := `select count(1) from system_schema.tables where keyspace_name=? and table_name = ?`
	row := c.session.Query(query, keyspace, tableName)
	tableExists := 0
	if err := row.Scan(&tableExists); err != nil {
		return nil, errors.Wrap(err, "failed to scan")
	}

	if tableExists == 0 && cassandraTableSchema.IsDeleted {
		return []string{}, nil
	} else if tableExists > 0 && cassandraTableSchema.IsDeleted {
		return []string{
			fmt.Sprintf(`drop table %s.%s`, keyspace, tableName),
		}, nil
	}

	if tableExists == 0 {
		// shortcut to just create it
		queries, err := CreateTableStatements(keyspace, tableName, cassandraTableSchema)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create table statement")
		}

		return queries, nil
	}

	statements := []string{}

	columnStatements, err := buildColumnStatements(c, tableName, cassandraTableSchema)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build column statements")
	}
	statements = append(statements, columnStatements...)

	propertiesStatements, err := buildPropertiesStatements(c, tableName, cassandraTableSchema)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build properties statements")
	}
	statements = append(statements, propertiesStatements...)

	return statements, nil
}

func buildPropertiesStatements(c *CassandraConnection, tableName string, cassandraTableSchema *schemasv1alpha4.CassandraTableSchema) ([]string, error) {
	query := `select bloom_filter_fp_chance, caching, comment, compaction, compression, crc_check_chance, dclocal_read_repair_chance
from system_schema.tables
where keyspace_name = ? and table_name = ?`
	row := c.session.Query(query, c.keyspace, tableName)

	currentProperties := schemasv1alpha4.CassandraTableProperties{}

	var bloomFilterFPChance float64
	var crcCheckChance float64
	var dcLocalReadRepairChance float64

	err := row.Scan(&bloomFilterFPChance, &currentProperties.Caching, &currentProperties.Comment,
		&currentProperties.Compaction, &currentProperties.Compression, &crcCheckChance,
		&dcLocalReadRepairChance)
	if err != nil {
		return nil, errors.Wrap(err, "failed to scan")
	}

	currentProperties.BloomFilterFPChance = fmt.Sprintf("%f", bloomFilterFPChance)
	currentProperties.CRCCheckChance = fmt.Sprintf("%f", crcCheckChance)
	currentProperties.DCLocalReadRepairChance = fmt.Sprintf("%f", dcLocalReadRepairChance)

	needsUpdating := []string{}

	if cassandraTableSchema.Properties != nil {
		if currentProperties.BloomFilterFPChance != cassandraTableSchema.Properties.BloomFilterFPChance {
			needsUpdating = append(needsUpdating, fmt.Sprintf("bloom_filter_fp_chance = %s", cassandraTableSchema.Properties.BloomFilterFPChance))
		}
		if !reflect.DeepEqual(currentProperties.Caching, cassandraTableSchema.Properties.Caching) {
			if cassandraTableSchema.Properties.Caching == nil {
				needsUpdating = append(needsUpdating, "caching = null")
			} else {
				b, err := json.Marshal(cassandraTableSchema.Properties.Caching)
				if err != nil {
					return nil, errors.Wrap(err, "failed to marshal caching property")
				}

				// TODO is there a better way? cassandra only accepts these maps with single quotes
				formatted := strings.ReplaceAll(string(b), `"`, `'`)
				needsUpdating = append(needsUpdating, fmt.Sprintf("caching = %s", formatted))
			}
		}
		if currentProperties.Comment != cassandraTableSchema.Properties.Comment {
			// TODO escape
			needsUpdating = append(needsUpdating, fmt.Sprintf("comment = '%s'", cassandraTableSchema.Properties.Comment))
		}
		if !reflect.DeepEqual(currentProperties.Compaction, cassandraTableSchema.Properties.Compaction) {
			if cassandraTableSchema.Properties.Compaction == nil {
				needsUpdating = append(needsUpdating, "compaction = null")
			} else {
				b, err := json.Marshal(cassandraTableSchema.Properties.Compaction)
				if err != nil {
					return nil, errors.Wrap(err, "failed to marshal compaction property")
				}

				// TODO is there a better way? cassandra only accepts these maps with single quotes
				formatted := strings.ReplaceAll(string(b), `"`, `'`)
				needsUpdating = append(needsUpdating, fmt.Sprintf("compaction = %s", formatted))
			}
		}
		if !reflect.DeepEqual(currentProperties.Compression, cassandraTableSchema.Properties.Compression) {
			if cassandraTableSchema.Properties.Compression == nil {
				needsUpdating = append(needsUpdating, "compression = null")
			} else {
				b, err := json.Marshal(cassandraTableSchema.Properties.Compression)
				if err != nil {
					return nil, errors.Wrap(err, "failed to marshal compression property")
				}

				// TODO is there a better way? cassandra only accepts these maps with single quotes
				formatted := strings.ReplaceAll(string(b), `"`, `'`)
				needsUpdating = append(needsUpdating, fmt.Sprintf("compression = %s", formatted))
			}
		}
		if currentProperties.CRCCheckChance != cassandraTableSchema.Properties.CRCCheckChance {
			// TODO escape
			needsUpdating = append(needsUpdating, fmt.Sprintf("crc_check_chance = '%s'", cassandraTableSchema.Properties.CRCCheckChance))
		}
		if currentProperties.DCLocalReadRepairChance != cassandraTableSchema.Properties.DCLocalReadRepairChance {
			// TODO escape
			needsUpdating = append(needsUpdating, fmt.Sprintf("dclocal_read_repair_chance = '%s'", cassandraTableSchema.Properties.DCLocalReadRepairChance))
		}
		if currentProperties.DefaultTTL != cassandraTableSchema.Properties.DefaultTTL {
			// TODO escape
			needsUpdating = append(needsUpdating, fmt.Sprintf("default_time_to_live = %d", *cassandraTableSchema.Properties.DefaultTTL))
		}
		if currentProperties.GCGraceSeconds != cassandraTableSchema.Properties.GCGraceSeconds {
			// TODO escape
			needsUpdating = append(needsUpdating, fmt.Sprintf("gc_grace_period_seconds = %d", *cassandraTableSchema.Properties.GCGraceSeconds))
		}
		if currentProperties.MaxIndexInterval != cassandraTableSchema.Properties.MaxIndexInterval {
			// TODO escape
			needsUpdating = append(needsUpdating, fmt.Sprintf("max_index_interval = %d", *cassandraTableSchema.Properties.MaxIndexInterval))
		}
		if currentProperties.MemtableFlushPeriodMS != cassandraTableSchema.Properties.MemtableFlushPeriodMS {
			// TODO escape
			needsUpdating = append(needsUpdating, fmt.Sprintf("memtable_flush_period_in_ms = %d", *cassandraTableSchema.Properties.MemtableFlushPeriodMS))
		}
		if currentProperties.MinIndexInterval != cassandraTableSchema.Properties.MinIndexInterval {
			// TODO escape
			needsUpdating = append(needsUpdating, fmt.Sprintf("min_index_interval = %d", *cassandraTableSchema.Properties.MinIndexInterval))
		}
		if currentProperties.ReadRepairChance != cassandraTableSchema.Properties.ReadRepairChance {
			// TODO escape
			needsUpdating = append(needsUpdating, fmt.Sprintf("read_repair_chance = '%s'", cassandraTableSchema.Properties.ReadRepairChance))
		}
		if currentProperties.SpeculativeRetry != cassandraTableSchema.Properties.SpeculativeRetry {
			// TODO escape
			needsUpdating = append(needsUpdating, fmt.Sprintf("speculative_retry = '%s'", cassandraTableSchema.Properties.SpeculativeRetry))
		}
	}

	if len(needsUpdating) == 0 {
		return []string{}, nil
	}

	return []string{
		fmt.Sprintf("alter table %s.%s with %s", c.keyspace, tableName, strings.Join(needsUpdating, " and ")),
	}, nil
}

func buildColumnStatements(c *CassandraConnection, tableName string, cassandraTableSchema *schemasv1alpha4.CassandraTableSchema) ([]string, error) {
	query := `select column_name, type from system_schema.columns where
keyspace_name = ? and table_name = ?`
	scanner := c.session.Query(query, c.keyspace, tableName).Iter().Scanner()

	alterAndDropStatements := []string{}
	foundColumnNames := []string{}

	for scanner.Next() {

		var columnName, columnType string

		if err := scanner.Scan(&columnName, &columnType); err != nil {
			return nil, errors.Wrap(err, "failed to scan")
		}

		foundColumnNames = append(foundColumnNames, columnName)

		existingColumn := types.Column{
			Name:     columnName,
			DataType: columnType,
		}

		columnStatement, err := AlterColumnStatements(c.keyspace, tableName, cassandraTableSchema.Columns, &existingColumn)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create alter column statement")
		}

		alterAndDropStatements = append(alterAndDropStatements, columnStatement...)
	}

	for _, desiredColumn := range cassandraTableSchema.Columns {
		isColumnPresent := false
		for _, foundColumn := range foundColumnNames {
			if foundColumn == desiredColumn.Name {
				isColumnPresent = true
			}
		}

		if !isColumnPresent {
			statement, err := InsertColumnStatement(c.keyspace, tableName, desiredColumn)
			if err != nil {
				return nil, errors.Wrap(err, "failed to create insert column statement")
			}

			alterAndDropStatements = append(alterAndDropStatements, statement)
		}
	}

	return alterAndDropStatements, nil
}

func DeployCassandraStatements(hosts []string, username string, password string, keyspace string, statements []string) error {
	c, err := Connect(hosts, username, password, keyspace)
	if err != nil {
		return err
	}
	defer c.Close()

	// execute
	if err := executeStatements(c, statements); err != nil {
		return err
	}

	return nil
}

func executeStatements(c *CassandraConnection, statements []string) error {
	for _, statement := range statements {
		if statement == "" {
			continue
		}
		fmt.Printf("Executing query %q\n", statement)
		if err := c.session.Query(statement); err != nil {
			return err.Context().Err()
		}
	}

	return nil
}
