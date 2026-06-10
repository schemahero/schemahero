package slackapp

import (
	"context"

	"github.com/pkg/errors"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
	schemasclientv1alpha4 "github.com/schemahero/schemahero/pkg/client/schemaheroclientset/typed/schemas/v1alpha4"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
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
		return false, err
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
			return false, errors.Errorf("migration %s/%s is %s", migration.Namespace, migration.Name, migration.Status.Phase)
		default:
			return false, nil
		}
	}

	return matched > 0, nil
}
