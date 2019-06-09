package mysql

import (
	"time"

	//	_ "github.com/lib/mysql"
	databasesclientv1alpha2 "github.com/schemahero/schemahero/pkg/client/schemaheroclientset/typed/databases/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func (m *MysqlConnection) CheckAlive(namespace string, instanceName string) (bool, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return false, err
	}
	databasesClient, err := databasesclientv1alpha2.NewForConfig(cfg)
	if err != nil {
		return false, err
	}
	instance, err := databasesClient.Databases(namespace).Get(instanceName, metav1.GetOptions{})
	if err != nil {
		return false, err
	}

	err = m.db.Ping()
	isConnected := err == nil
	if err != nil {
		instance.Status.IsConnected = false
	} else {
		instance.Status.IsConnected = true
		instance.Status.LastPing = time.Now().Format(time.RFC3339)
	}

	_, err = databasesClient.Databases(namespace).Update(instance)
	if err != nil {
		return isConnected, err
	}

	return isConnected, nil
}
