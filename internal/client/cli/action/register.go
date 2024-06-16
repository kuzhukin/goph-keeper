package action

import (
	"crypto/aes"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/kuzhukin/goph-keeper/internal/client/cli/args"
	"github.com/kuzhukin/goph-keeper/internal/client/gophcrypto"
	"github.com/kuzhukin/goph-keeper/internal/client/storage"
	"github.com/kuzhukin/goph-keeper/internal/client/transport"
	"github.com/urfave/cli/v2"
)

func RegisterAction(
	s storage.UserStorage,
	client *transport.Client,
) func(ctx *cli.Context) error {
	return func(ctx *cli.Context) error {
		login := args.GetLogin(ctx)
		password := args.GetPassword(ctx)

		cryptoKey, err := generateKey()
		if err != nil {
			return err
		}

		user := &storage.User{Login: login, IsActive: false, CryptoKey: cryptoKey}

		crypto, err := gophcrypto.New(user.CryptoKey)
		if err != nil {
			return err
		}

		encryptedPassword := crypto.Encrypt([]byte(password))
		user.Password = encryptedPassword

		token, err := client.RegisterUser(login, encryptedPassword)
		if err != nil {
			return fmt.Errorf("registration on server failed with error: %w", err)
		}

		err = s.Register(ctx.Context, login, encryptedPassword, token, base64.RawStdEncoding.EncodeToString(cryptoKey))
		if err != nil {
			return err
		}

		return nil
	}

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
