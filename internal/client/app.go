package client

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/kuzhukin/goph-keeper/internal/client/config"
	"github.com/urfave/cli/v2"
)

const configFileName = "client_config.yaml"

type Application struct {
	cli    cli.App
	client *Client
}

func NewApplication() (*Application, error) {
	var client *Client
	var conf *config.Config
	var err error

	initClientFunc := func(ctx *cli.Context) error {
		if ctx.Bool("help") {
			cli.ShowAppHelpAndExit(ctx, 0)
		}

		if ctx.Command.Name == "config" {
			return nil
		}

		conf, err = config.ReadConfig(configFileName)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				fmt.Println("Use config command for authentification. See gophkeep-client config --help")
			} else {
				fmt.Printf("Unknown error: %s\n", err)
			}

			return err
		}

		client = newClient(conf)

		return nil
	}

	cli := cli.App{
		Name:     "gophkeep",
		Version:  "v1.0.0",
		Compiled: time.Now(),
		Usage:    "Use for send your data to gokeep server",
		Commands: []*cli.Command{
			{
				Name:        "config",
				Usage:       "Client configuration",
				Description: "Add configuration to ~/.goph-keeper/client_config.yaml",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "login",
						Usage: "User's login",
					},
					&cli.StringFlag{
						Name:  "password",
						Usage: "User's password",
					},
					&cli.StringFlag{
						Name:  "server-url",
						Usage: "Pair of server's IP:port",
					},
				},
				Action: func(ctx *cli.Context) error {
					params := map[string]string{}
					for _, flag := range ctx.FlagNames() {
						value := ctx.String(flag)
						if len(value) != 0 {
							params[flag] = value
						}
					}

					if len(params) == 0 {
						cli.ShowAppHelpAndExit(ctx, 1)
					}

					if err = config.UpdateConfig(configFileName, params); err != nil {
						fmt.Printf("update config error: %s\n", err)
					}

					return err
				},
			},
			{
				Name:    "create",
				Aliases: []string{"c"},
				Usage:   "Send new file to server",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "file",
						Aliases: []string{"f"},
						Usage:   "Path of creating file",
					},
				},
				Description: "Read file and send it to server",
				Before:      initClientFunc,
				Action: func(ctx *cli.Context) error {
					filename := ctx.String("file")
					if len(filename) == 0 {
						cli.ShowAppHelpAndExit(ctx, 1)
					}

					err := client.CreateFile(filename)
					if err != nil {
						fmt.Println(err)
					}

					return err
				},
			},
			{
				Name:    "get",
				Aliases: []string{"g"},
				Usage:   "Download file from server",
				Before:  initClientFunc,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "file",
						Aliases: []string{"f"},
						Usage:   "path to file",
					},
					&cli.StringFlag{
						Name:    "output-dir",
						Aliases: []string{"o"},
						Usage:   "Output directory",
					},
				},
				Action: func(ctx *cli.Context) error {
					filename := ctx.String("file")
					if len(filename) == 0 {
						cli.ShowAppHelpAndExit(ctx, 1)
					}

					resp, err := client.GetFile(filename)
					if err != nil {
						fmt.Println(err)

						return err
					}

					outputDir := ctx.String("output-dir")
					if len(outputDir) == 0 {
						outputDir = "."
					}

					_ = os.MkdirAll(outputDir, 0700)

					outputFilePath := filepath.Join(outputDir, filename)

					err = os.WriteFile(outputFilePath, []byte(resp.Data), 0600)
					if err != nil {
						return err
					}

					return nil
				},
			},
			{
				Name:    "update",
				Aliases: []string{"u"},
				Usage:   "Update existed file on server",
				Before:  initClientFunc,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "file",
						Aliases: []string{"f"},
						Usage:   "path to file",
					},
					&cli.BoolFlag{
						Name:  "force",
						Usage: "rewrite file on the server",
					},
				},
			},
			{
				Name:    "delete",
				Aliases: []string{"d"},
				Usage:   "Delete file on server",
				Before:  initClientFunc,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "file",
						Aliases: []string{"f"},
						Usage:   "path to file",
					},
				},
			},
		},
	}

	return &Application{cli: cli, client: client}, nil
}

func (a *Application) Run() error {
	if err := a.cli.Run(os.Args); err != nil {
		return err
	}

	return nil
}
