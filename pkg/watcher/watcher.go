package watcher

import (
	"fmt"
	"time"

	"github.com/schemahero/schemahero/pkg/schemahero/postgres"

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

	db, err := postgres.Connect(w.Viper.GetString("uri"))
	if err != nil {
		return err
	}

	for {
		if err := db.CheckAlive(); err != nil {
			return err
		}

		time.Sleep(time.Second * 10)
	}
}
