package client

import (
	"crypto/aes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/kuzhukin/goph-keeper/internal/client/config"
	"github.com/kuzhukin/goph-keeper/internal/client/gophcrypto"
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

		user.CryptoKey, err = base64.RawStdEncoding.DecodeString(string(user.CryptoKey))
		if err != nil {
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
			a.makeGetDataCmd(),
			a.makeListCmd(),
			a.makeUpdateFileCmd(),
			a.makeDeleteBinaryDataCmd(),
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

	cryptoKey, err := generateKey()
	if err != nil {
		return nil
	}

	crypto, err := gophcrypto.New(cryptoKey)
	if err != nil {
		return err
	}

	encryptedPassword := crypto.Encrypt([]byte(password), vecFromUser(crypto, a.user))

	err = a.storage.Register(login, encryptedPassword, base64.RawStdEncoding.EncodeToString(cryptoKey))
	if err != nil {
		return err
	}

	err = a.client.Register(login, encryptedPassword)
	if err != nil {
		return err
	}

	return nil
}

func vecFromUser(crypto *gophcrypto.Cryptographer, u *storage.User) []byte {
	s := sha256.Sum256(u.CryptoKey)
	vec := s[:crypto.VecLen()]

	return vec
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
	r, err := a.readDataFromFileArg(ctx)
	if err != nil {
		return fmt.Errorf("read data from file, err=%w", err)
	}

	rev, err := a.storage.Save(a.user, r)
	if err != nil {
		return err
	}

	r.Revision = rev

	err = a.client.Upload(a.user, r)
	if err != nil {
		return err
	}

	fmt.Printf("Data from file=%s is saves", r.Name)

	return nil
}

func (a *Application) makeGetDataCmd() *cli.Command {
	return &cli.Command{
		Name:    "get",
		Aliases: []string{"g"},
		// FIXME: по факту файл должен тянутся из локальной базы, для стягивания с сервера должна быть другая команда
		Usage:  "Download file from server",
		Before: a.checkConfig,
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
		Action: a.getDataCmdHandler,
	}
}

func (a *Application) getDataCmdHandler(ctx *cli.Context) error {
	filename := getFileArg(ctx)

	file, err := a.client.Download(a.user, filename)
	if err != nil {
		fmt.Println(err)

		return err
	}

	fmt.Println(file.Data)

	decryptedData, err := a.decryptUserData([]byte(file.Data))
	if err != nil {
		return err
	}

	fmt.Println(string(decryptedData))

	return nil
}

func (a *Application) makeListCmd() *cli.Command {
	return &cli.Command{
		Name:    "list",
		Aliases: []string{"l"},
		Usage:   "Print local data names and revisions",
		Before:  a.checkConfig,
		Action:  a.listCmdHandler,
	}
}

func (a *Application) listCmdHandler(ctx *cli.Context) error {
	records, err := a.storage.List(a.user)
	if err != nil {
		return err
	}

	if len(records) == 0 {
		fmt.Println("Records isn't exist")
	}

	for _, r := range records {
		fmt.Printf("\t%s (%d)\n", r.Name, r.Revision)
	}

	return nil
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
		Action: a.updateDataCmdHandler,
	}
}

func (a *Application) updateDataCmdHandler(ctx *cli.Context) error {
	r, err := a.readDataFromFileArg(ctx)
	if err != nil {
		return err
	}

	rev, err := a.storage.Save(a.user, r)
	if err != nil {
		return nil
	}

	r.Revision = rev

	err = a.client.Upload(a.user, r)
	if err != nil {
		return err
	}

	return nil
}

func (a *Application) makeDeleteBinaryDataCmd() *cli.Command {
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

// func (a *Application) delete

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

func (a *Application) readDataFromFileArg(ctx *cli.Context) (*storage.Record, error) {
	filename := getFileArg(ctx)

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	encryptedData, err := a.encryptUserData(data)
	if err != nil {
		return nil, err
	}

	r := &storage.Record{Name: filename, Data: string(encryptedData), Revision: 1}

	return r, nil
}

func (a *Application) decryptUserData(data []byte) ([]byte, error) {
	crypto, err := gophcrypto.New([]byte(a.user.CryptoKey))
	if err != nil {
		return nil, err
	}

	return crypto.Decrypt(data, vecFromUser(crypto, a.user))
}

func (a *Application) encryptUserData(data []byte) ([]byte, error) {
	crypto, err := gophcrypto.New([]byte(a.user.CryptoKey))
	if err != nil {
		return nil, err
	}

	return []byte(crypto.Encrypt(data, vecFromUser(crypto, a.user))), nil
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

func generateKey() ([]byte, error) {
	const keyLenDefault = aes.BlockSize

	data, err := generateRandom(keyLenDefault)
	if err != nil {
		return nil, nil
	}

	return data, nil
}

func generateRandom(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}
