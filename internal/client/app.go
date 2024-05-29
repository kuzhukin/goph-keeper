package client

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/kuzhukin/goph-keeper/internal/client/config"
	"github.com/kuzhukin/goph-keeper/internal/client/storage"
	"github.com/urfave/cli/v2"
)

const configFileName = "client_config.yaml"

type Application struct {
	cli     cli.App
	client  *Client
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

		app.storage, err = storage.StartNewDbStorage(conf.Database)
		if err != nil {
			return nil, err
		}

		user, err := app.storage.User()
		if err != nil && !errors.Is(err, storage.ErrNotActiveOrRegistredUsers) {
			return nil, err
		}

		app.user = user
		app.client = newClient(conf)
	}

	return app, nil
}

func (a *Application) initCLI() {
	a.cli = cli.App{
		Name:     "gophkeep",
		Version:  "v1.0.0",
		Compiled: time.Now(),
		Usage:    "Use for send your data to gokeep server",
		Commands: []*cli.Command{
			a.makeConfigCmd(),
			a.makeRegisterCmd(),
			a.makeCreateFileCmd(),
			a.makeGetFileCmd(),
			a.makeUpdateFileCmd(),
			a.makeDeleteFileCmd(),
		},
	}
}

func (a *Application) makeConfigCmd() *cli.Command {
	return &cli.Command{
		Name:        "config",
		Usage:       "Client configuration",
		Description: "Add configuration to ~/.goph-keeper/client_config.yaml",
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
		Before: a.checkConfig,
		Action: a.configCmdHandler,
	}
}

func (a *Application) configCmdHandler(ctx *cli.Context) error {
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

	err := config.UpdateConfig(configFileName, params)
	if err != nil {
		fmt.Printf("update config error: %s\n", err)
	}

	return err
}

func (a *Application) makeRegisterCmd() *cli.Command {
	return &cli.Command{
		Name:        "register",
		Usage:       "Registrates user in system",
		Description: "You shoud register before using application",
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
		Action: a.registerCmdHandler,
	}
}

func (a *Application) registerCmdHandler(ctx *cli.Context) error {
	login := getLoginArg(ctx)
	password := getPasswordArg(ctx)
	encryptedPassword := encodePassword(password)

	cryptoKey, err := generateKey()
	if err != nil {
		return nil
	}

	cryptoKeyStr := string(cryptoKey)

	err = a.storage.Register(login, encryptedPassword, cryptoKeyStr)
	if err != nil {
		return err
	}

	err = a.client.Register(login, encryptedPassword)
	if err != nil {
		return err
	}

	return nil
}

func encodePassword(password string) string {
	h := sha256.New()
	h.Write([]byte(password))

	return base64.RawStdEncoding.EncodeToString(h.Sum(nil))
}

func (a *Application) makeCreateFileCmd() *cli.Command {
	return &cli.Command{
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
		Before:      a.checkConfig,
		Action:      a.createFileCmdHander,
	}
}

func (a *Application) createFileCmdHander(ctx *cli.Context) error {
	filename := getFileArg(ctx)

	err := a.client.CreateFile(a.user.Login, a.user.Password, filename)
	if err != nil {
		fmt.Println(err)
	}

	return err
}

func (a *Application) makeGetFileCmd() *cli.Command {
	return &cli.Command{
		Name:    "get",
		Aliases: []string{"g"},
		Usage:   "Download file from server",
		Before:  a.checkConfig,
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
		Action: a.getFileCmdHandler,
	}
}

func (a *Application) getFileCmdHandler(ctx *cli.Context) error {
	filename := getFileArg(ctx)

	file, err := a.client.GetFile(a.user.Login, a.user.Password, filename)
	if err != nil {
		fmt.Println(err)

		return err
	}

	outputDir := getOutputDirArg(ctx)

	return file.Save(outputDir)
}

func (a *Application) makeUpdateFileCmd() *cli.Command {
	return &cli.Command{
		Name:    "update",
		Aliases: []string{"u"},
		Usage:   "Update existed file on server",
		Before:  a.checkConfig,
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
	}
}

func (a *Application) makeDeleteFileCmd() *cli.Command {
	return &cli.Command{
		Name:    "delete",
		Aliases: []string{"d"},
		Usage:   "Delete file on server",
		Before:  a.checkConfig,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "file",
				Aliases: []string{"f"},
				Usage:   "path to file",
			},
		},
	}
}

// func (a *Application) makeEditFileCmd() *cli.Command {
// 	return &cli.Command{
// 		// TODO: должен запускать текстовый редактор с файлом и изменять ревизию
// 		// {
// 		// 	Name:  "edit",
// 		// 	Usage: "Edit file",
// 		// 	Flags: []cli.Flag{
// 		// 		&cli.StringFlag{
// 		// 			Name:    "file",
// 		// 			Aliases: []string{"f"},
// 		// 			Usage:   "path to file",
// 		// 		},
// 		// 	},
// 		// 	Action: func(ctx *cli.Context) error {

// 		// 	},
// 		// },
// 	}
// }

func (a *Application) checkConfig(ctx *cli.Context) error {
	if a.config == nil {
		fmt.Println("client isn't configured")

		cli.ShowAppHelpAndExit(ctx, 1)
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

func getFileArg(ctx *cli.Context) string {
	filename := ctx.String("file")
	if len(filename) == 0 {
		cli.ShowAppHelpAndExit(ctx, 1)
	}

	return filename
}

func getLoginArg(ctx *cli.Context) string {
	value := ctx.String("login")
	if len(value) == 0 {
		cli.ShowAppHelpAndExit(ctx, 1)
	}

	return value
}

func getPasswordArg(ctx *cli.Context) string {
	value := ctx.String("password")
	if len(value) == 0 {
		cli.ShowAppHelpAndExit(ctx, 1)
	}

	return value
}

func getOutputDirArg(ctx *cli.Context) string {
	outputDir := ctx.String("output-dir")
	if len(outputDir) == 0 {
		outputDir = "."
	}

	return outputDir
}

func generateKey() (string, error) {
	const keyLenDefault = 32

	data, err := generateRandom(keyLenDefault)
	if err != nil {
		return "", nil
	}

	return string(data), nil
}

func generateRandom(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}
