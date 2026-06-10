package slackapp

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRenderWorkflowTemplate(t *testing.T) {
	template := strings.Join([]string{
		`release_tag="${PRODUCTION_RELEASE_TAG:-${{ inputs.tag || github.ref_name }}}"`,
		`release_sha="${PRODUCTION_RELEASE_SHA:-${{ inputs.sha || github.sha }}}"`,
		`previous_sha="${PRODUCTION_PREVIOUS_SHA:-${{ inputs.previous-sha }}}"`,
		`previous_tag_name="${PRODUCTION_PREVIOUS_TAG_NAME:-${{ inputs.previous-tag-name }}}"`,
	}, "\n")
	templateFile, err := os.CreateTemp("", "workflow-template-*.yml")
	require.NoError(t, err)
	defer os.Remove(templateFile.Name())
	_, err = templateFile.WriteString(template)
	require.NoError(t, err)
	require.NoError(t, templateFile.Close())

	renderedPath, err := renderWorkflowTemplate(templateFile.Name(), Release{
		Tag:             "v2026.06.10-0",
		SHA:             "release-sha",
		PreviousSHA:     "previous-sha",
		PreviousTagName: "v2026.06.09-0",
	})
	require.NoError(t, err)
	defer os.Remove(renderedPath)

	rendered, err := os.ReadFile(renderedPath)
	require.NoError(t, err)
	require.Contains(t, string(rendered), `release_tag="v2026.06.10-0"`)
	require.Contains(t, string(rendered), `release_sha="release-sha"`)
	require.Contains(t, string(rendered), `previous_sha="previous-sha"`)
	require.Contains(t, string(rendered), `previous_tag_name="v2026.06.09-0"`)
}

func TestNewDepotDispatcherFromEnvRequiresRepo(t *testing.T) {
	t.Setenv("DEPOT_CI_DISPATCH_TOKEN", "token")
	t.Setenv("DEPOT_WORKFLOW_TEMPLATE_PATH", "/tmp/workflow.yml")
	t.Setenv("DEPOT_REPO", "")

	dispatcher, ok := NewDepotDispatcherFromEnv()

	require.False(t, ok)
	require.Nil(t, dispatcher)
}

func TestNewDepotDispatcherFromEnvUsesConfiguredRepo(t *testing.T) {
	t.Setenv("DEPOT_CI_DISPATCH_TOKEN", "token")
	t.Setenv("DEPOT_WORKFLOW_TEMPLATE_PATH", "/tmp/workflow.yml")
	t.Setenv("DEPOT_REPO", "example/repo")

	dispatcher, ok := NewDepotDispatcherFromEnv()

	require.True(t, ok)
	require.Equal(t, "example/repo", dispatcher.Repo)
}
