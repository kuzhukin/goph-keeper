package action

import (
	"errors"
	"fmt"

	"github.com/kuzhukin/goph-keeper/internal/client/cli/args"
	"github.com/kuzhukin/goph-keeper/internal/client/storage"
	"github.com/kuzhukin/goph-keeper/internal/client/storage/sqlstorage"
	"github.com/kuzhukin/goph-keeper/internal/client/transport"
	"github.com/urfave/cli/v2"
)

func CreateSecretActionHandler(
	user *storage.User,
	storage storage.SecretStorage,
	client *transport.Client,
) func(ctx *cli.Context) error {
	return func(ctx *cli.Context) error {
		secret, err := args.GetSecret(ctx)
		if err != nil {
			return fmt.Errorf("can't create secret: %w", err)
		}

		cryptedSecret, err := storage.CreateSecret(ctx.Context, user, secret)
		if err != nil && !errors.Is(err, sqlstorage.ErrAlreadyExist) {
			return err
		}

		if err = client.CreateSecret(user.Token, secret.Name, cryptedSecret); err != nil {
			return err
		}

		return nil
	}

}

func GetSecretActionHandler(
	user *storage.User,
	storage storage.SecretStorage,
	client *transport.Client,
) func(ctx *cli.Context) error {
	return func(ctx *cli.Context) error {
		key := ctx.String("name")
		if len(key) == 0 {
			return errors.New("bad secret's name")
		}

		secret, err := storage.GetSecret(ctx.Context, user, key)
		if err != nil {
			return err
		}

		fmt.Printf("\tname: %s; key: %s; value: %s\n", secret.Name, secret.Key, secret.Value)

		return nil
	}
}

func DeleteSecretActionHandler(
	user *storage.User,
	storage storage.SecretStorage,
	client *transport.Client,
) func(ctx *cli.Context) error {
	return func(ctx *cli.Context) error {
		key := ctx.String("name")
		if len(key) == 0 {
			return errors.New("bad secret's name")
		}

		// we are firstly deleting data on the server
		if err := client.DeleteSecret(user.Token, key); err != nil {
			return err
		}

		if err := storage.DeleteSecret(ctx.Context, user, key); err != nil {
			return err
		}

		return nil
	}

}
