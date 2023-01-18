/*
Copyright 2019 The SchemaHero Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha4

type TimescaleDBHypertable struct {
	TimeColumnName *string `json:"timeColumnName,omitempty" yaml:"timeColumnName,omitempty"`

	PartitioningColumn    *string  `json:"partitioningColumn,omitempty" yaml:"partitioningColumn,omitempty"`
	NumberPartitions      *int     `json:"numberPartitions,omitempty" yaml:"numberPartitions,omitempty"`
	ChunkTimeInterval     *string  `json:"chunkTimeInterval,omitempty" yaml:"chunkTimeInterval,omitempty"`
	CreateDefaultIndexes  *bool    `json:"createDefaultIndexes,omitempty" yaml:"createDefaultIndexes,omitempty"`
	IfNotExists           *bool    `json:"ifNotExists,omitempty" yaml:"ifNotExists,omitempty"`
	PartitioningFunc      *string  `json:"partitioningFunc,omitempty" yaml:"partitioningFunc,omitempty"`
	AssociatedSchemaName  *string  `json:"associatedSchemaName,omitempty" yaml:"associatedSchemaName,omitempty"`
	AssociatedTablePrefix *string  `json:"associatedTablePrefix,omitempty" yaml:"associatedTablePrefix,omitempty"`
	MigrateData           *bool    `json:"migrateData,omitempty" yaml:"migrateData,omitempty"`
	TimePartitioningFunc  *string  `json:"timePartitioningFunc,omitempty" yaml:"timePartitioningFunc,omitempty"`
	ReplicationFactor     *int     `json:"replicationFactor,omitempty" yaml:"replicationFactor,omitempty"`
	DataNodes             []string `json:"dataNodes,omitempty" yaml:"dataNodes,omitempty"`

	Compression *TimescaleDBCompression `json:"compression,omitempty" yaml:"compression,omitempty"`
	Retention   *TimescaleDBRetention   `json:"retention,omitempty" yaml:"retention,omitempty"`
}

type TimescaleDBCompression struct {
	SegmentBy *string `json:"segmentBy" yaml:"segmentBy"`
	Interval  *string `json:"interval" yaml:"interval"`
}

type TimescaleDBRetention struct {
	Interval string `json:"interval" yaml:"interval"`
}

type TimescaleDBTableSchema struct {
	PrimaryKey  []string                     `json:"primaryKey,omitempty" yaml:"primaryKey,omitempty"`
	ForeignKeys []*PostgresqlTableForeignKey `json:"foreignKeys,omitempty" yaml:"foreignKeys,omitempty"`
	Indexes     []*PostgresqlTableIndex      `json:"indexes,omitempty" yaml:"indexes,omitempty"`
	Columns     []*PostgresqlTableColumn     `json:"columns,omitempty" yaml:"columns,omitempty"`
	IsDeleted   bool                         `json:"isDeleted,omitempty" yaml:"isDeleted,omitempty"`
	Triggers    []*PostgresqlTableTrigger    `json:"triggers,omitempty" yaml:"triggers,omitempty"`
	Hypertable  *TimescaleDBHypertable       `json:"hypertable,omitempty" yaml:"hypertable,omitempty"`
}

type TimescaleDBViewSchema struct {
	IsContinuousAggregate *bool  `json:"isContinuousAggregate,omitempty" yaml:"isContinuousAggregate,omitempty"`
	WithNoData            *bool  `json:"withNoData,omitempty" yaml:"withNoData,omitempty"`
	Query                 string `json:"query,omitempty" yaml:"query,omitempty"`
	IsDeleted             bool   `json:"isDeleted,omitempty" yaml:"isDeleted,omitempty"`
}
