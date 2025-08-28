package postgres

import "testing"

func Test_stripOIDClass(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{
			name:  "basic",
			value: `'11'::integer`,
			want:  "11",
		},
		{
			name:  "empty",
			value: `''::character varying`,
			want:  "",
		},
		{
			name:  "identity",
			value: `testing`,
			want:  "testing",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stripOIDClass(tt.value); got != tt.want {
				t.Errorf("stripOIDClass() = %v, want %v", got, tt.want)
			}
		})
	}
}
