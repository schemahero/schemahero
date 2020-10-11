package postgres

import (
	"testing"

	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_triggerCreateStatement(t *testing.T) {
	condition := "OLD.balance IS DISTINCT FROM NEW.balance"

	tests := []struct {
		name              string
		trigger           *schemasv1alpha4.PostgresqlTableTrigger
		tableName         string
		expectedStatement string
	}{
		{
			name: "after insert",
			trigger: &schemasv1alpha4.PostgresqlTableTrigger{
				Name: "tt",
				Events: []string{
					"after insert",
				},
				ForEachRow:       &trueValue,
				ExecuteProcedure: "fn()",
			},
			tableName:         "a",
			expectedStatement: `create trigger "tt" after insert on "a" for each row execute procedure fn()`,
		},
		{
			name: "with when",
			trigger: &schemasv1alpha4.PostgresqlTableTrigger{
				Name: "tt",
				Events: []string{
					"before update",
				},
				ForEachRow:       &trueValue,
				Condition:        &condition,
				ExecuteProcedure: "fn()",
			},
			tableName:         "a",
			expectedStatement: `create trigger "tt" before update on "a" for each row when (OLD.balance IS DISTINCT FROM NEW.balance) execute procedure fn()`,
		},
		{
			name: "after insert or update",
			trigger: &schemasv1alpha4.PostgresqlTableTrigger{
				Name: "tt",
				Events: []string{
					"after insert",
					"after update",
				},
				ForEachRow:       &trueValue,
				ExecuteProcedure: "fn()",
			},
			tableName:         "a",
			expectedStatement: `create trigger "tt" after insert or update on "a" for each row execute procedure fn()`,
		},
		{
			name: "before insert constraint trigger",
			trigger: &schemasv1alpha4.PostgresqlTableTrigger{
				Name:              "tt",
				ConstraintTrigger: &trueValue,
				Events: []string{
					"before insert",
				},
				ForEachStatement: &trueValue,
				ExecuteProcedure: "fn()",
			},
			tableName:         "a",
			expectedStatement: `create constraint trigger "tt" before insert on "a" for each statement execute procedure fn()`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := require.New(t)

			actual, err := triggerCreateStatement(test.trigger, test.tableName)
			req.NoError(err)

			assert.Equal(t, test.expectedStatement, actual)
		})
	}
}

func Test_triggerEventSyntax(t *testing.T) {
	tests := []struct {
		name              string
		trigger           *schemasv1alpha4.PostgresqlTableTrigger
		expectedStatement string
	}{
		{
			name: "after insert",
			trigger: &schemasv1alpha4.PostgresqlTableTrigger{
				Events: []string{
					"after insert",
				},
			},
			expectedStatement: `after insert`,
		},
		{
			name: "after insert and truncate",
			trigger: &schemasv1alpha4.PostgresqlTableTrigger{
				Name: "t",
				Events: []string{
					"after insert",
					"after truncate",
				},
			},
			expectedStatement: `after insert or truncate`,
		},
		{
			name: "after insert and update of one column",
			trigger: &schemasv1alpha4.PostgresqlTableTrigger{
				Events: []string{
					"after insert",
					`after update of "c"`,
				},
			},
			expectedStatement: `after insert or update of "c"`,
		},
		{
			name: "instead of insert on",
			trigger: &schemasv1alpha4.PostgresqlTableTrigger{
				Events: []string{
					`instead of insert`,
				},
			},
			expectedStatement: `instead of insert`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := require.New(t)

			actual, err := triggerEvent(test.trigger)
			req.NoError(err)

			assert.Equal(t, test.expectedStatement, actual)
		})
	}
}
