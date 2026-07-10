package v1alpha4

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_IsEmpty(t *testing.T) {
	tests := []struct {
		name   string
		input  ValueOrValueFrom
		expect bool
	}{
		{
			name: "static value",
			input: ValueOrValueFrom{
				Value: "test",
			},
			expect: false,
		},
		{
			name: "from a secret",
			input: ValueOrValueFrom{
				ValueFrom: &ValueFrom{
					SecretKeyRef: &SecretKeyRef{
						Name: "test",
						Key:  "test",
					},
				},
			},
			expect: false,
		},
		{
			name:   "empty",
			input:  ValueOrValueFrom{},
			expect: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := test.input.IsEmpty()
			assert.Equal(t, test.expect, actual)
		})
	}
}

func TestGetValueRejectsAmbiguousSources(t *testing.T) {
	database := &Database{}
	_, err := database.getValueFromValueOrValueFrom(context.Background(), "postgres", &ValueOrValueFrom{
		Value: "postgres://literal",
		ValueFrom: &ValueFrom{
			Doppler: validDopplerConfig(),
		},
	})
	assert.EqualError(t, err, "exactly one of value, secretKeyRef, vault, ssm, or doppler must be configured")
}
