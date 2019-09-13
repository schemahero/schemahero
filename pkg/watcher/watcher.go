package watcher

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/schemahero/schemahero/pkg/database/interfaces"
	"github.com/schemahero/schemahero/pkg/database/mysql"
	"github.com/schemahero/schemahero/pkg/database/postgres"
	"github.com/spf13/viper"
)

type Watcher struct {
	Viper *viper.Viper
}

func NewWatcher() *Watcher {
	return &Watcher{
		Viper: viper.GetViper(),
	}
}

func (w *Watcher) RunSync() error {
	fmt.Printf("connecting to %s\n", w.Viper.GetString("uri"))

	var conn interfaces.SchemaHeroDatabaseConnection

	if w.Viper.GetString("driver") == "postgres" {
		c, err := postgres.Connect(w.Viper.GetString("uri"))
		if err != nil {
			return errors.Wrap(err, "postgres connect")
		}

		conn = c
	} else if w.Viper.GetString("driver") == "mysql" {
		c, err := mysql.Connect(w.Viper.GetString("uri"))
		if err != nil {
			return errors.Wrap(err, "mysql connect")
		}

		conn = c
	}

	for {
		if _, err := conn.CheckAlive(w.Viper.GetString("namespace"), w.Viper.GetString("instance")); err != nil {
			return errors.Wrap(err, "check alive")
		}

		time.Sleep(time.Second * 10)
	}
}
