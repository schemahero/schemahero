package postgres

import (
	"context"
	"time"

	_ "github.com/lib/pq"
	databasesclientv1alpha3 "github.com/schemahero/schemahero/pkg/client/schemaheroclientset/typed/databases/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func (p *PostgresConnection) CheckAlive(ctx context.Context, namespace string, instanceName string) (bool, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return false, err
	}
	databasesClient, err := databasesclientv1alpha3.NewForConfig(cfg)
	if err != nil {
		return false, err
	}
	instance, err := databasesClient.Databases(namespace).Get(ctx, instanceName, metav1.GetOptions{})
	if err != nil {
		return false, err
	}

	err = p.db.Ping()
	isConnected := err == nil
	if err != nil {
		instance.Status.IsConnected = false
	} else {
		instance.Status.IsConnected = true
		instance.Status.LastPing = time.Now().Format(time.RFC3339)
	}

	_, err = databasesClient.Databases(namespace).Update(ctx, instance, metav1.UpdateOptions{})
	if err != nil {
		return isConnected, err
	}

	return isConnected, nil
}
