package main

import (
	"fmt"

	"github.com/kuzhukin/goph-keeper/internal/client"
	"github.com/kuzhukin/goph-keeper/internal/zlog"
)

func main() {
	defer func() {
		_ = zlog.Logger().Sync()
	}()

	if err := run(); err != nil {
		fmt.Println("error:", err)
	}
}

func run() error {
	app, err := client.NewApplication()
	if err != nil {
		return err
	}

	err = app.Run()
	if err != nil {
		fmt.Println("failed: ", err)
	}

	return nil
}
