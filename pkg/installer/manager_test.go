package installer

import (
	"strings"
	"testing"
)

func TestManagerYAMLOmitsStatus(t *testing.T) {
	got, err := managerYAML("schemahero-system")
	if err != nil {
		t.Fatalf("managerYAML() error = %v", err)
	}

	manifest := string(got)
	if !strings.Contains(manifest, "kind: StatefulSet") {
		t.Fatalf("managerYAML() should render a StatefulSet, got:\n%s", manifest)
	}

	if strings.Contains(manifest, "\nstatus:") {
		t.Fatalf("managerYAML() should not render status, got:\n%s", manifest)
	}
}
