package watcher

import (
	"fmt"

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

	fmt.Printf("%#v\n", db)
	return nil
}
