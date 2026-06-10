package slackapp

import (
	"context"
	stderrors "errors"

	"github.com/pkg/errors"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	schemasclientv1alpha4 "github.com/schemahero/schemahero/pkg/client/schemaheroclientset/typed/schemas/v1alpha4"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

var (
	errNoReleaseMigrations = stderrors.New("no matching release migrations")
	errReleaseMigrationEnd = stderrors.New("release migration reached terminal phase")
)

func releaseFromMigration(migration *schemasv1alpha4.Migration) (Release, bool) {
	annotations := migration.GetAnnotations()
	release := Release{
		Tag:             annotations[ReleaseTagAnnotation],
		SHA:             annotations[ReleaseSHAAnnotation],
		PreviousSHA:     annotations[PreviousSHAAnnotation],
		PreviousTagName: annotations[PreviousTagNameAnnotation],
	}
	if release.Tag == "" || release.SHA == "" || release.PreviousSHA == "" || release.PreviousTagName == "" {
		return Release{}, false
	}
	return release, true
}

func releaseMigrationsExecuted(ctx context.Context, schemasClient schemasclientv1alpha4.SchemasV1alpha4Interface, release Release, namespace string) (bool, error) {
	migrations, err := schemasClient.Migrations(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labels.Everything().String(),
	})
	if err != nil {
		return false, errors.Wrap(err, "failed to list release migrations")
	}

	matched := 0
	for _, migration := range migrations.Items {
		annotations := migration.GetAnnotations()
		if annotations[ReleaseTagAnnotation] != release.Tag || annotations[ReleaseSHAAnnotation] != release.SHA {
			continue
		}
		matched++
		switch migration.Status.Phase {
		case schemasv1alpha4.Executed:
			continue
		case schemasv1alpha4.Rejected, schemasv1alpha4.Invalid:
			return false, errors.Wrapf(errReleaseMigrationEnd, "migration %s/%s is %s", migration.Namespace, migration.Name, migration.Status.Phase)
		default:
			return false, nil
		}
	}

	if matched == 0 {
		return false, errors.Wrapf(errNoReleaseMigrations, "release %s at %s", release.Tag, release.SHA)
	}
	return true, nil
}

func releaseReadinessErrorIsTerminal(err error) bool {
	return stderrors.Is(err, errNoReleaseMigrations) || stderrors.Is(err, errReleaseMigrationEnd)
}
