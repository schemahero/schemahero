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

type CassandraColumn struct {
	Name     string `json:"name" yaml:"name"`
	Type     string `json:"type" yaml:"type"`
	IsStatic *bool  `json:"isStatic,omitempty" yaml:"isStatic,omitempty"`
}

type CassandraClusteringOrder struct {
	Column       string `json:"column" yaml:"column"`
	IsDescending *bool  `json:"isDescending,omitempty" yaml:"isDescending,omitempty"`
}

type CassandraTableProperties struct {
	BloomFilterFPChance     string            `json:"bloomFilterFPChance,omitempty" yaml:"bloomFilterFPChance,omitempty"`
	Caching                 map[string]string `json:"caching,omitempty" yaml:"caching,omitempty"`
	Comment                 string            `json:"comment,omitempty" yaml:"comment,omitempty"`
	Compaction              map[string]string `json:"compaction,omitempty" yaml:"compaction,omitempty"`
	Compression             map[string]string `json:"compression,omitempty" yaml:"compression,omitempty"`
	CRCCheckChance          string            `json:"crcCheckChance,omitempty" yaml:"crcCheckChance,omitempty"`
	DCLocalReadRepairChance string            `json:"dcLocalReadRepairChance,omitempty" yaml:"dcLocalReadRepairChance,omitempty"`
	DefaultTTL              *int              `json:"defaultTTL,omitempty" yaml:"defaultTTL,omitempty"`
	GCGraceSeconds          *int              `json:"gcGraceSeconds,omitempty" yaml:"gcGraceSeconds,omitempty"`
	MaxIndexInterval        *int              `json:"maxIndexInterval,omitempty" yaml:"maxIndexInterval,omitempty"`
	MemtableFlushPeriodMS   *int              `json:"memtableFlushPeriodMs,omitempty" yaml:"memtableFlushPeriodMs,omitempty"`
	MinIndexInterval        *int              `json:"minIndexInterval,omitempty" yaml:"minIndexInterval,omitempty"`
	ReadRepairChance        string            `json:"readRepairChance,omitempty" yaml:"readRepairChance,omitempty"`
	SpeculativeRetry        string            `json:"speculativeRetry,omitempty" yaml:"speculativeRetry,omitempty"`
}

type CassandraTableSchema struct {
	IsDeleted       bool                      `json:"isDeleted,omitempty" yaml:"isDeleted,omitempty"`
	PrimaryKey      [][]string                `json:"primaryKey,omitempty" yaml:"primaryKey,omitempty"`
	ClusteringOrder *CassandraClusteringOrder `json:"clusteringOrder,omitempty" yaml:"clusteringOrder,omitempty"`
	Columns         []*CassandraColumn        `json:"columns,omitempty" yaml:"columns,omitempty"`

	Properties *CassandraTableProperties `json:"properties,omitempty" yaml:"properties,omitempty"`
}

type CassandraField struct {
	Name string `json:"name" yaml:"name"`
	Type string `json:"type" yaml:"type"`
}

type CassandraDataTypeSchema struct {
	IsDeleted bool              `json:"isDeleted,omitempty" yaml:"isDeleted,omitempty"`
	Fields    []*CassandraField `json:"fields,omitempty" yaml:"fields,omitempty"`
}

// UnmarshalYAML implements custom YAML unmarshaling to handle zero values properly
func (p *CassandraTableProperties) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Use a map to capture all fields including zero values
	var raw map[string]interface{}
	if err := unmarshal(&raw); err != nil {
		return err
	}

	// Handle string fields
	if v, ok := raw["bloomFilterFPChance"]; ok {
		if str, ok := v.(string); ok {
			p.BloomFilterFPChance = str
		}
	}
	if v, ok := raw["comment"]; ok {
		if str, ok := v.(string); ok {
			p.Comment = str
		}
	}
	if v, ok := raw["crcCheckChance"]; ok {
		if str, ok := v.(string); ok {
			p.CRCCheckChance = str
		}
	}
	if v, ok := raw["dcLocalReadRepairChance"]; ok {
		if str, ok := v.(string); ok {
			p.DCLocalReadRepairChance = str
		}
	}
	if v, ok := raw["readRepairChance"]; ok {
		if str, ok := v.(string); ok {
			p.ReadRepairChance = str
		}
	}
	if v, ok := raw["speculativeRetry"]; ok {
		if str, ok := v.(string); ok {
			p.SpeculativeRetry = str
		}
	}

	// Handle map fields
	if v, ok := raw["caching"]; ok {
		if m, ok := v.(map[interface{}]interface{}); ok {
			p.Caching = make(map[string]string)
			for k, val := range m {
				p.Caching[k.(string)] = val.(string)
			}
		}
	}
	if v, ok := raw["compaction"]; ok {
		if m, ok := v.(map[interface{}]interface{}); ok {
			p.Compaction = make(map[string]string)
			for k, val := range m {
				p.Compaction[k.(string)] = val.(string)
			}
		}
	}
	if v, ok := raw["compression"]; ok {
		if m, ok := v.(map[interface{}]interface{}); ok {
			p.Compression = make(map[string]string)
			for k, val := range m {
				p.Compression[k.(string)] = val.(string)
			}
		}
	}

	// Handle integer fields - explicitly check for presence
	if v, ok := raw["defaultTTL"]; ok {
		if i, ok := v.(int); ok {
			p.DefaultTTL = &i
		}
	}
	if v, ok := raw["gcGraceSeconds"]; ok {
		if i, ok := v.(int); ok {
			p.GCGraceSeconds = &i
		}
	}
	if v, ok := raw["maxIndexInterval"]; ok {
		if i, ok := v.(int); ok {
			p.MaxIndexInterval = &i
		}
	}
	if v, ok := raw["memtableFlushPeriodMs"]; ok {
		if i, ok := v.(int); ok {
			p.MemtableFlushPeriodMS = &i
		}
	}
	if v, ok := raw["minIndexInterval"]; ok {
		if i, ok := v.(int); ok {
			p.MinIndexInterval = &i
		}
	}

	return nil
}
