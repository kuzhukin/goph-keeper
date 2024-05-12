package main

import (
	"github.com/kuzhukin/goph-keeper/internal/client"
	"github.com/kuzhukin/goph-keeper/internal/zlog"
)

func main() {
	defer func() {
		_ = zlog.Logger().Sync()
	}()

	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	app, err := client.NewApplication()
	if err != nil {
		return err
	}

	app.Run()

	return nil
}
