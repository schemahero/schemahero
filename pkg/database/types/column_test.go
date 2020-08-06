package types

import "testing"

func TestBoolsEqual(t *testing.T) {
	falseValue := false
	trueValue := true
	tests := []struct {
		name string
		a    *bool
		b    *bool
		want bool
	}{
		{
			name: "false false",
			a:    &falseValue,
			b:    &falseValue,
			want: true,
		},
		{
			name: "true true",
			a:    &trueValue,
			b:    &trueValue,
			want: true,
		},
		{
			name: "false true",
			a:    &falseValue,
			b:    &trueValue,
			want: false,
		},
		{
			name: "true false",
			a:    &trueValue,
			b:    &falseValue,
			want: false,
		},
		{
			name: "nil nil",
			a:    nil,
			b:    nil,
			want: true,
		},
		{
			name: "nil false",
			a:    nil,
			b:    &falseValue,
			want: true,
		},
		{
			name: "false nil",
			a:    &falseValue,
			b:    nil,
			want: true,
		},
		{
			name: "nil true",
			a:    nil,
			b:    &trueValue,
			want: false,
		},
		{
			name: "true nil",
			a:    &trueValue,
			b:    nil,
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BoolsEqual(tt.a, tt.b); got != tt.want {
				t.Errorf("BoolsEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}
