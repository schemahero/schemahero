package slackapp

import (
	"context"
	"testing"

	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	testclient "github.com/schemahero/schemahero/pkg/client/schemaheroclientset/fake"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestReleaseMigrationsExecutedRequiresMatchingMigrations(t *testing.T) {
	client := testclient.NewSimpleClientset(plannedMigration("other-release")).SchemasV1alpha4()

	ready, err := releaseMigrationsExecuted(context.Background(), client, Release{
		Tag: "v2099.01.01-1",
		SHA: "2222222222222222222222222222222222222222",
	}, "default")

	require.False(t, ready)
	require.Error(t, err)
	require.True(t, releaseReadinessErrorIsTerminal(err))
}

func TestReleaseMigrationsExecutedReturnsReadyWhenMatchesAreExecuted(t *testing.T) {
	migration := plannedMigration("executed")
	migration.Status.Phase = schemasv1alpha4.Executed
	client := testclient.NewSimpleClientset(migration).SchemasV1alpha4()

	ready, err := releaseMigrationsExecuted(context.Background(), client, Release{
		Tag: migration.Annotations[ReleaseTagAnnotation],
		SHA: migration.Annotations[ReleaseSHAAnnotation],
	}, "default")

	require.NoError(t, err)
	require.True(t, ready)
}

func TestReleaseMigrationsExecutedReturnsTerminalErrorForRejectedMigration(t *testing.T) {
	migration := plannedMigration("rejected")
	migration.Status.Phase = schemasv1alpha4.Rejected
	client := testclient.NewSimpleClientset(migration).SchemasV1alpha4()

	ready, err := releaseMigrationsExecuted(context.Background(), client, Release{
		Tag: migration.Annotations[ReleaseTagAnnotation],
		SHA: migration.Annotations[ReleaseSHAAnnotation],
	}, metav1.NamespaceDefault)

	require.False(t, ready)
	require.Error(t, err)
	require.True(t, releaseReadinessErrorIsTerminal(err))
}
