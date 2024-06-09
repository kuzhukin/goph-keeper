package client

import (
	"crypto/aes"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"time"
	"unicode"

	"github.com/kuzhukin/goph-keeper/internal/client/config"
	"github.com/kuzhukin/goph-keeper/internal/client/gophcrypto"
	"github.com/kuzhukin/goph-keeper/internal/client/storage"
	"github.com/kuzhukin/goph-keeper/internal/client/storage/sqlstorage"
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

		dbStorage, err := sqlstorage.StartNewDbStorage(conf.Database)
		if err != nil {
			return nil, err
		}

		app.storage = dbStorage

		user, err := app.storage.GetActive()
		if err != nil && !errors.Is(err, sqlstorage.ErrNotActiveOrRegistredUsers) {
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
		Name:         "gophkeep",
		Version:      "v1.0.0",
		Compiled:     time.Now(),
		BashComplete: cli.DefaultAppComplete,
		Usage:        "Use for send your data to gokeep server",
		Commands: []*cli.Command{
			a.makeConfigCmd(),
			a.makeRegisterCmd(),
			a.makeDataCmd(),
		},
	}
}

func (a *Application) makeWalletCmd() *cli.Command {
	return &cli.Command{
		Name:   "wallet",
		Usage:  "Operations with bank's cards",
		Before: a.checkConfig,
		Subcommands: []*cli.Command{
			a.makeCreateCardCmd(),
			a.makeDeleteCardCmd(),
			a.makeListCardCmd(),
		},
	}
}

func (a *Application) makeCreateCardCmd() *cli.Command {
	return &cli.Command{
		Name:  "create",
		Usage: "Create new card",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "number"},
			&cli.StringFlag{Name: "expiration", Usage: fmt.Sprintf("Data in format: %s", storage.ExpirationFormat)},
			&cli.StringFlag{Name: "cvv"},
			&cli.StringFlag{Name: "owner"},
		},
		Action: a.createCardCmdHandler,
	}
}

func (a *Application) createCardCmdHandler(ctx *cli.Context) error {
	card, err := makeBankCardFromArgs(ctx)
	if err != nil {
		return fmt.Errorf("Can't create card: %s", err)
	}

	return a.storage.CreateCard(a.user, card)
}

func makeBankCardFromArgs(ctx *cli.Context) (*storage.BankCard, error) {
	number, ok := validateCardNumber(ctx.String("number"))
	if !ok {
		return nil, errors.New("bad card's number")
	}

	exp, ok := validateExpDate(ctx.String("expiration"))
	if !ok {
		return nil, errors.New("bad expiration date")
	}

	cvv, ok := validateCvvCode(ctx.String("cvv"))
	if !ok {
		return nil, errors.New("bad card's cvv")
	}

	owner, ok := validateCardOwner("owner")
	if !ok {
		return nil, errors.New("bad card owner name")
	}

	return &storage.BankCard{
		Number:     number,
		ExpiryDate: exp,
		Owner:      owner,
		CvvCode:    cvv,
	}, nil
}

// format: "IVAN PETROV"
func validateCardOwner(owner string) (string, bool) {
	findSpace := false

	for _, r := range owner {
		if unicode.IsLetter(r) {
			continue
		}

		if unicode.IsSpace(r) {
			if findSpace {
				return "", false
			}

			findSpace = true
			continue
		}

		return "", false
	}

	if !findSpace {
		return "", false
	}

	return owner, true
}

func validateExpDate(date string) (time.Time, bool) {
	exp, err := time.Parse(storage.ExpirationFormat, date)
	if err != nil {
		return time.Time{}, false
	}

	return exp, true
}

func validateCvvCode(cvvCode string) (string, bool) {
	if len(cvvCode) != 3 {
		return "", false
	}

	for _, r := range cvvCode {
		if !unicode.IsDigit(r) {
			return "", false
		}
	}

	return cvvCode, true
}

func validateCardNumber(number string) (string, bool) {
	// number validation
	validatedNumber := make([]byte, 0, 16)
	for _, r := range number {
		if r < 255 && unicode.IsDigit(r) {
			validatedNumber = append(validatedNumber, byte(r))
		} else if unicode.IsSpace(r) {
			continue
		}

		return "", false
	}

	if len(validatedNumber) != 16 {
		return "", false
	}

	return string(validatedNumber), true
}

func (a *Application) makeDeleteCardCmd() *cli.Command {
	return &cli.Command{
		Name:  "delete",
		Usage: "Delete card",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "number"},
		},
		Action: a.deleteCardCmdHandler,
	}
}

func (a *Application) deleteCardCmdHandler(ctx *cli.Context) error {
	number, ok := validateCardNumber(ctx.String("number"))
	if !ok {
		return errors.New("bad card number")
	}

	// TODO: add deletion on server

	return a.storage.DeleteCard(a.user, number)
}

func (a *Application) listCardCmdHandler(ctx *cli.Context) error {
	list, err := a.storage.ListCard(a.user)
	if err != nil {
		return err
	}

	for _, card := range list {
		fmt.Printf("%s %s %v (%s)", card.Owner, card.Number, card.ExpiryDate.Format(storage.ExpirationFormat), card.CvvCode)
	}

	return nil
}

func (a *Application) makeListCardCmd() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List with all user cards",
	}
}

func (a *Application) makeDataCmd() *cli.Command {
	return &cli.Command{
		Name:   "data",
		Usage:  "Operations with text or binary data",
		Before: a.checkConfig,
		Subcommands: []*cli.Command{
			a.makeCreateCmd(),
			a.makeGetCmd(),
			a.makeListCmd(),
			a.makeUpdateCmd(),
			a.makeDeleteCmd(),
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

	a.user = &storage.User{Login: login, IsActive: false, CryptoKey: cryptoKey}

	crypto, err := gophcrypto.New(a.user.CryptoKey)
	if err != nil {
		return err
	}

	encryptedPassword := crypto.Encrypt([]byte(password))
	a.user.Password = encryptedPassword

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

func (a *Application) makeCreateCmd() *cli.Command {
	return &cli.Command{
		Name:  "create",
		Usage: "Send new file to server",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "file",
				Aliases: []string{"f"},
				Usage:   "Path of creating file",
			},
		},
		Description: "Read file and send it to server",
		Action:      a.createCmdHander,
	}
}

func (a *Application) createCmdHander(ctx *cli.Context) error {
	r, err := a.readDataFromFileArg(ctx)
	if err != nil {
		return fmt.Errorf("read data from file, err=%w", err)
	}

	rev, err := a.storage.SaveData(a.user, r)
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

func (a *Application) makeGetCmd() *cli.Command {
	return &cli.Command{
		Name: "get",
		// FIXME: по факту файл должен тянутся из локальной базы, для стягивания с сервера должна быть другая команда
		Usage: "Download file from server",
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
		Name:   "list",
		Usage:  "Print local data names and revisions",
		Action: a.listCmdHandler,
	}
}

func (a *Application) listCmdHandler(ctx *cli.Context) error {
	records, err := a.storage.ListData(a.user)
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

func (a *Application) makeUpdateCmd() *cli.Command {
	return &cli.Command{
		Name:  "update",
		Usage: "Update existed data on server",
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
		Action: a.updateCmdHandler,
	}
}

func (a *Application) updateCmdHandler(ctx *cli.Context) error {
	r, err := a.readDataFromFileArg(ctx)
	if err != nil {
		return err
	}

	rev, err := a.storage.SaveData(a.user, r)
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

func (a *Application) makeDeleteCmd() *cli.Command {
	return &cli.Command{
		Name:    "delete",
		Aliases: []string{"d"},
		Usage:   "Delete data on server",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "file",
				Aliases: []string{"f"},
				Usage:   "path to file",
			},
		},
		Action: a.deleteBinaryDataCmdHandler,
	}
}

func (a *Application) deleteBinaryDataCmdHandler(ctx *cli.Context) error {
	r, err := a.readDataFromFileArg(ctx)
	if err != nil {
		return err
	}

	return a.storage.DeleteData(a.user, r)
}

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
	crypto, err := gophcrypto.New(a.user.CryptoKey)
	if err != nil {
		return nil, err
	}

	return crypto.Decrypt(data)
}

func (a *Application) encryptUserData(data []byte) ([]byte, error) {
	crypto, err := gophcrypto.New([]byte(a.user.CryptoKey))
	if err != nil {
		return nil, err
	}

	return []byte(crypto.Encrypt(data)), nil
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
