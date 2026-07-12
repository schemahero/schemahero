package installer

import "testing"

func TestGeneratedTableCRDEnablesStatusSubresource(t *testing.T) {
	tableCRD := tablesCRDV1()

	for _, version := range tableCRD.Spec.Versions {
		if !version.Served {
			continue
		}
		if version.Subresources == nil || version.Subresources.Status == nil {
			t.Fatalf("served Table CRD version %q does not enable the status subresource", version.Name)
		}
	}
}
