package cli

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/kuzhukin/goph-keeper/internal/client/cli/action"
	"github.com/kuzhukin/goph-keeper/internal/client/config"
	"github.com/kuzhukin/goph-keeper/internal/client/storage"
	"github.com/kuzhukin/goph-keeper/internal/client/storage/sqlstorage"
	"github.com/kuzhukin/goph-keeper/internal/client/transport"
	"github.com/urfave/cli/v2"
)

const configFileName = "client_config.yaml"

type Application struct {
	cli     cli.App
	client  *transport.Client
	user    *storage.User
	storage storage.Storage
	config  *config.Config
}

func NewApplication() (*Application, error) {
	app := &Application{}

	conf, err := config.ReadConfig(configFileName)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Println("Use config command for authentification. See gophkeep-client config --help")

			return nil, err
		} else {
			fmt.Printf("Unknown error: %s\n", err)

			return nil, err
		}
	}

	app.initCLI()

	if conf != nil {
		app.config = conf

		dbStorage, err := sqlstorage.StartNewDbStorage(conf.Database)
		if err != nil {
			return nil, err
		}

		app.storage = dbStorage

		user, err := app.storage.GetActive(context.Background())
		if err != nil {
			if !errors.Is(err, sqlstorage.ErrNotActiveOrRegistredUsers) {
				return nil, err
			}
		}

		if user != nil {
			user.CryptoKey, err = base64.RawStdEncoding.DecodeString(string(user.CryptoKey))
			if err != nil {
				return nil, err
			}

			app.user = user
		}

		app.client = transport.NewClient(conf)
	}

	return app, nil
}

func (a *Application) initCLI() {
	a.cli = cli.App{
		Name:         "gophkeep",
		Version:      "v1.0.0",
		Compiled:     time.Now(),
		BashComplete: cli.DefaultAppComplete,
		Usage:        "Use for send your data to gokeep server",
		Commands: []*cli.Command{
			a.makeConfigCmd(),
			a.makeRegisterCmd(),
			a.makeDataCmd(),
			a.makeWalletCmd(),
			a.makeSecretCmd(),
		},
	}
}

func (a *Application) makeSecretCmd() *cli.Command {
	return &cli.Command{
		Name:         "secret",
		Usage:        "Operations with bank's cards",
		Before:       a.checkConfig,
		BashComplete: cli.DefaultAppComplete,
		Subcommands: []*cli.Command{
			a.makeCreateSecretCmd(),
			a.makeDeleteSecretCmd(),
			a.makeGetSecretCmd(),
		},
	}
}

func (a *Application) makeCreateSecretCmd() *cli.Command {
	return &cli.Command{
		Name:         "create",
		Usage:        "Create new secret",
		BashComplete: cli.DefaultAppComplete,
		Before:       a.checkConfig,
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "name"},
			&cli.StringFlag{Name: "key"},
			&cli.StringFlag{Name: "value"},
		},
		Action: action.CreateSecretActionHandler(a.user, a.storage, a.client),
	}
}

func (a *Application) makeDeleteSecretCmd() *cli.Command {
	return &cli.Command{
		Name:         "delete",
		Usage:        "Create new secret",
		BashComplete: cli.DefaultAppComplete,
		Before:       a.checkConfig,
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "name"},
		},
		Action: action.DeleteSecretActionHandler(a.user, a.storage, a.client),
	}
}

func (a *Application) makeGetSecretCmd() *cli.Command {
	return &cli.Command{
		Name:         "get",
		Usage:        "Create new secret",
		BashComplete: cli.DefaultAppComplete,
		Before:       a.checkConfig,
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "name"},
		},
		Action: action.GetSecretActionHandler(a.user, a.storage, a.client),
	}
}

func (a *Application) makeWalletCmd() *cli.Command {
	return &cli.Command{
		Name:         "wallet",
		Usage:        "Operations with bank's cards",
		BashComplete: cli.DefaultAppComplete,
		Before:       a.checkConfig,
		Subcommands: []*cli.Command{
			a.makeCreateCardCmd(),
			a.makeDeleteCardCmd(),
			a.makeListCardCmd(),
		},
	}
}

func (a *Application) makeCreateCardCmd() *cli.Command {
	return &cli.Command{
		Name:         "create",
		Usage:        "Create new card",
		BashComplete: cli.DefaultAppComplete,
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "number"},
			&cli.StringFlag{Name: "expiration", Usage: fmt.Sprintf("Data in format: %s", storage.ExpirationFormat)},
			&cli.StringFlag{Name: "cvv"},
			&cli.StringFlag{Name: "owner"},
		},
		Action: action.CreateCardActionHandler(a.user, a.storage, a.client),
	}
}

func (a *Application) makeDeleteCardCmd() *cli.Command {
	return &cli.Command{
		Name:         "delete",
		Usage:        "Delete card",
		BashComplete: cli.DefaultAppComplete,
		Before:       a.checkConfig,
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "number"},
		},
		Action: action.DeleteCardActionHandler(a.user, a.storage, a.client),
	}
}

func (a *Application) makeListCardCmd() *cli.Command {
	return &cli.Command{
		Name:         "list",
		Usage:        "List with all user cards",
		BashComplete: cli.DefaultAppComplete,
		Before:       a.checkConfig,
		Action:       action.ListCardActionHandler(a.user, a.storage, a.client),
	}
}

func (a *Application) makeDataCmd() *cli.Command {
	return &cli.Command{
		Name:         "data",
		Usage:        "Operations with text or binary data",
		BashComplete: cli.DefaultAppComplete,
		Before:       a.checkConfig,
		Subcommands: []*cli.Command{
			a.makeCreateDataCmd(),
			a.makeGetDataCmd(),
			a.makeListDataCmd(),
			a.makeUpdateDataCmd(),
			a.makeDeleteDataCmd(),
		},
	}
}

func (a *Application) makeConfigCmd() *cli.Command {
	return &cli.Command{
		Name:        "config",
		Usage:       "Client configuration",
		Description: "Add configuration to ~/.goph-keeper/client_config.yaml",
		Before:      a.checkConfig,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "server-url",
				Usage: "Pair of server's IP:port",
			},
			&cli.StringFlag{
				Name:  "database-name",
				Usage: "Database's file's name",
			},
		},
		Action: action.ConfigAction(configFileName),
	}
}

func (a *Application) makeRegisterCmd() *cli.Command {
	return &cli.Command{
		Name:         "register",
		Usage:        "Registrates user in system",
		Description:  "You shoud register before using application",
		BashComplete: cli.DefaultAppComplete,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "login",
				Usage: "User's login",
			},
			&cli.StringFlag{
				Name:  "password",
				Usage: "User's password",
			},
		},
		Action: action.RegisterAction(a.storage, a.client),
	}
}

func (a *Application) makeCreateDataCmd() *cli.Command {
	return &cli.Command{
		Name:         "create",
		Usage:        "Send new file to server",
		BashComplete: cli.DefaultAppComplete,
		Before:       a.checkConfig,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "file",
				Aliases: []string{"f"},
				Usage:   "Path of creating file",
			},
		},
		Description: "Read file and send it to server",
		Action:      action.CreateDataAction(a.user, a.storage, a.client),
	}
}

func (a *Application) makeGetDataCmd() *cli.Command {
	return &cli.Command{
		Name:         "get",
		Usage:        "Download file from server",
		BashComplete: cli.DefaultAppComplete,
		Before:       a.checkConfig,
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
		Action: action.GetDataCmdHandler(a.user, a.storage, a.client),
	}
}

func (a *Application) makeListDataCmd() *cli.Command {
	return &cli.Command{
		Name:   "list",
		Usage:  "Print local data names and revisions",
		Before: a.checkConfig,
		Action: action.ListDataAction(a.user, a.storage, a.client),
	}
}

func (a *Application) makeUpdateDataCmd() *cli.Command {
	return &cli.Command{
		Name:         "update",
		Usage:        "Update existed data on server",
		BashComplete: cli.DefaultAppComplete,
		Before:       a.checkConfig,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "file",
				Aliases: []string{"f"},
				Usage:   "path to file",
			},
			&cli.BoolFlag{
				Name:  "force",
				Usage: "update date",
			},
		},
		Action: action.UpdateAction(a.user, a.storage, a.client),
	}
}

func (a *Application) makeDeleteDataCmd() *cli.Command {
	return &cli.Command{
		Name:         "delete",
		Aliases:      []string{"d"},
		Usage:        "Delete data on server",
		BashComplete: cli.DefaultAppComplete,
		Before:       a.checkConfig,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "file",
				Aliases: []string{"f"},
				Usage:   "path to file",
			},
		},
		Action: action.DeleteBinaryDataAction(a.user, a.storage, a.client),
	}
}

func (a *Application) checkConfig(ctx *cli.Context) error {
	if a.config == nil {
		fmt.Println("client isn't configured")

		cli.ShowAppHelpAndExit(ctx, 1)
	}

	if a.user == nil {
		fmt.Println("User isn't registred")
		cli.ShowCommandHelpAndExit(ctx, "register", 1)
	}

	if a.client == nil {
		fmt.Println("Need register before using goph-keeper client. Use --help for more information.")
		cli.ShowCommandHelpAndExit(ctx, "register", 1)
	}

	return nil
}

func (a *Application) Run() error {
	if err := a.cli.Run(os.Args); err != nil {
		return err
	}

	return nil
}
