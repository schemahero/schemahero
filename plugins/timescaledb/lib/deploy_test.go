package timescaledb

import (
	"testing"

	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	"github.com/stretchr/testify/assert"
)

func Test_policyDiffers(t *testing.T) {
	trueVar := true
	falseVar := false
	one := 1
	two := 2

	tests := []struct {
		name     string
		desired  *schemasv1alpha4.TimescaleDBViewRefreshPolicy
		current  *currentRefreshPolicy
		expected bool
	}{
		{
			name:     "both nil",
			desired:  nil,
			current:  nil,
			expected: false,
		},
		{
			name:     "desired nil, current exists",
			desired:  nil,
			current:  &currentRefreshPolicy{startOffset: "1 day"},
			expected: true,
		},
		{
			name:     "desired exists, current nil",
			desired:  &schemasv1alpha4.TimescaleDBViewRefreshPolicy{StartOffset: "1 day", EndOffset: "1 hour", ScheduleInterval: "1 hour"},
			current:  nil,
			expected: true,
		},
		{
			name: "identical policies",
			desired: &schemasv1alpha4.TimescaleDBViewRefreshPolicy{
				StartOffset:      "1 day",
				EndOffset:        "1 hour",
				ScheduleInterval: "1 hour",
			},
			current: &currentRefreshPolicy{
				startOffset:      "1 day",
				endOffset:        "1 hour",
				scheduleInterval: "1 hour",
			},
			expected: false,
		},
		{
			name: "startOffset differs",
			desired: &schemasv1alpha4.TimescaleDBViewRefreshPolicy{
				StartOffset:      "2 days",
				EndOffset:        "1 hour",
				ScheduleInterval: "1 hour",
			},
			current: &currentRefreshPolicy{
				startOffset:      "1 day",
				endOffset:        "1 hour",
				scheduleInterval: "1 hour",
			},
			expected: true,
		},
		{
			name: "scheduleInterval differs",
			desired: &schemasv1alpha4.TimescaleDBViewRefreshPolicy{
				StartOffset:      "1 day",
				EndOffset:        "1 hour",
				ScheduleInterval: "2 hours",
			},
			current: &currentRefreshPolicy{
				startOffset:      "1 day",
				endOffset:        "1 hour",
				scheduleInterval: "1 hour",
			},
			expected: true,
		},
		{
			name: "timezone differs",
			desired: &schemasv1alpha4.TimescaleDBViewRefreshPolicy{
				StartOffset:      "1 day",
				EndOffset:        "1 hour",
				ScheduleInterval: "1 hour",
				Timezone:         "UTC",
			},
			current: &currentRefreshPolicy{
				startOffset:      "1 day",
				endOffset:        "1 hour",
				scheduleInterval: "1 hour",
			},
			expected: true,
		},
		{
			name: "includeTieredData differs",
			desired: &schemasv1alpha4.TimescaleDBViewRefreshPolicy{
				StartOffset:       "1 day",
				EndOffset:         "1 hour",
				ScheduleInterval:  "1 hour",
				IncludeTieredData: &trueVar,
			},
			current: &currentRefreshPolicy{
				startOffset:      "1 day",
				endOffset:        "1 hour",
				scheduleInterval: "1 hour",
			},
			expected: true,
		},
		{
			name: "bucketsPerBatch differs",
			desired: &schemasv1alpha4.TimescaleDBViewRefreshPolicy{
				StartOffset:      "1 day",
				EndOffset:        "1 hour",
				ScheduleInterval: "1 hour",
				BucketsPerBatch:  &one,
			},
			current: &currentRefreshPolicy{
				startOffset:      "1 day",
				endOffset:        "1 hour",
				scheduleInterval: "1 hour",
			},
			expected: true,
		},
		{
			name: "maxBatchesPerExecution differs",
			desired: &schemasv1alpha4.TimescaleDBViewRefreshPolicy{
				StartOffset:            "1 day",
				EndOffset:              "1 hour",
				ScheduleInterval:       "1 hour",
				MaxBatchesPerExecution: &one,
			},
			current: &currentRefreshPolicy{
				startOffset:      "1 day",
				endOffset:        "1 hour",
				scheduleInterval: "1 hour",
			},
			expected: true,
		},
		{
			name: "refreshNewestFirst differs",
			desired: &schemasv1alpha4.TimescaleDBViewRefreshPolicy{
				StartOffset:        "1 day",
				EndOffset:          "1 hour",
				ScheduleInterval:   "1 hour",
				RefreshNewestFirst: &falseVar,
			},
			current: &currentRefreshPolicy{
				startOffset:      "1 day",
				endOffset:        "1 hour",
				scheduleInterval: "1 hour",
			},
			expected: true,
		},
		{
			name: "optional fields match empty current",
			desired: &schemasv1alpha4.TimescaleDBViewRefreshPolicy{
				StartOffset:      "1 day",
				EndOffset:        "1 hour",
				ScheduleInterval: "1 hour",
			},
			current: &currentRefreshPolicy{
				startOffset:      "1 day",
				endOffset:        "1 hour",
				scheduleInterval: "1 hour",
			},
			expected: false,
		},
		{
			name: "all optional fields match",
			desired: &schemasv1alpha4.TimescaleDBViewRefreshPolicy{
				StartOffset:            "1 day",
				EndOffset:              "1 hour",
				ScheduleInterval:       "1 hour",
				Timezone:               "UTC",
				IncludeTieredData:      &trueVar,
				BucketsPerBatch:        &one,
				MaxBatchesPerExecution: &two,
				RefreshNewestFirst:     &falseVar,
			},
			current: &currentRefreshPolicy{
				startOffset:              "1 day",
				endOffset:                "1 hour",
				scheduleInterval:         "1 hour",
				timezone:                 "UTC",
				includeTieredData:        "true",
				bucketsPerBatch:          "1",
				maxBatchesPerExecution:   "2",
				refreshNewestFirst:       "false",
			},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, policyDiffers(test.desired, test.current))
		})
	}
}

func Test_boolPtrString(t *testing.T) {
	trueVar := true
	falseVar := false

	assert.Equal(t, "", boolPtrString(nil))
	assert.Equal(t, "true", boolPtrString(&trueVar))
	assert.Equal(t, "false", boolPtrString(&falseVar))
}

func Test_intPtrString(t *testing.T) {
	one := 1

	assert.Equal(t, "", intPtrString(nil))
	assert.Equal(t, "1", intPtrString(&one))
}
