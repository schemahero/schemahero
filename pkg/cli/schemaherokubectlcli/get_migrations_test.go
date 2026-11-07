package schemaherokubectlcli

import (
	"testing"

	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
)

// TestIsPendingMigration verifies that only migrations that are neither executed nor rejected are pending
func TestIsPendingMigration(t *testing.T) {
	tests := []struct {
		name   string
		status schemasv1alpha4.MigrationStatus
		want   bool
	}{
		{name: "planned", status: schemasv1alpha4.MigrationStatus{PlannedAt: 1}, want: true},
		{name: "approved", status: schemasv1alpha4.MigrationStatus{PlannedAt: 1, ApprovedAt: 2}, want: true},
		{name: "executed", status: schemasv1alpha4.MigrationStatus{PlannedAt: 1, ApprovedAt: 2, ExecutedAt: 3}, want: false},
		{name: "rejected", status: schemasv1alpha4.MigrationStatus{PlannedAt: 1, RejectedAt: 2}, want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := isPendingMigration(schemasv1alpha4.Migration{Status: test.status})
			if got != test.want {
				t.Errorf("isPendingMigration() = %v, want %v", got, test.want)
			}
		})
	}
}
