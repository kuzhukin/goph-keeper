package action

import (
	"errors"
	"fmt"
	"os"

	"github.com/kuzhukin/goph-keeper/internal/client/cli/args"
	"github.com/kuzhukin/goph-keeper/internal/client/gophcrypto"
	"github.com/kuzhukin/goph-keeper/internal/client/storage"
	"github.com/kuzhukin/goph-keeper/internal/client/storage/sqlstorage"
	"github.com/kuzhukin/goph-keeper/internal/client/transport"
	"github.com/urfave/cli/v2"
)

func CreateDataAction(
	user *storage.User,
	s storage.DataStorage,
	client *transport.Client,
) func(ctx *cli.Context) error {
	return func(ctx *cli.Context) error {
		r, err := readDataFromFileArg(ctx, user)
		if err != nil {
			return fmt.Errorf("read data from file, err=%w", err)
		}

		err = s.CreateData(ctx.Context, user, r)
		if err != nil && !errors.Is(err, sqlstorage.ErrAlreadyExist) {
			return err
		}

		r.Revision = 1

		err = client.UploadBinaryData(user, r)
		if err != nil {
			return err
		}

		fmt.Printf("Data from file=%s is saved\ns", r.Name)

		return nil
	}
}

func GetDataCmdHandler(
	user *storage.User,
	s storage.DataStorage,
	client *transport.Client,
) func(ctx *cli.Context) error {
	return func(ctx *cli.Context) error {
		filename := args.GetFileArg(ctx)

		file, err := client.DownloadBinaryData(user, filename)
		if err != nil {
			return err
		}

		decryptedData, err := decryptUserData(user, []byte(file.Data))
		if err != nil {
			return err
		}

		fmt.Println(string(decryptedData))

		return nil
	}
}

func ListDataAction(
	user *storage.User,
	s storage.DataStorage,
	client *transport.Client,
) func(ctx *cli.Context) error {
	return func(ctx *cli.Context) error {
		records, err := s.ListData(ctx.Context, user)
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
}

func UpdateAction(
	user *storage.User,
	s storage.DataStorage,
	client *transport.Client,
) func(ctx *cli.Context) error {
	return func(ctx *cli.Context) error {
		r, err := readDataFromFileArg(ctx, user)
		if err != nil {
			return err
		}

		rev, needUpload, err := s.UpdateData(ctx.Context, user, r)
		if err != nil {
			return err
		}

		if needUpload {
			r.Revision = rev

			err = client.UpdateBinaryData(user, r)
			if err != nil {
				return err
			}
		} else {
			fmt.Println("Nothing for updating")
		}

		return nil
	}
}

func DeleteBinaryDataAction(
	user *storage.User,
	s storage.DataStorage,
	client *transport.Client,
) func(ctx *cli.Context) error {
	return func(ctx *cli.Context) error {
		r, err := readDataFromFileArg(ctx, user)
		if err != nil {
			return err
		}

		if err = s.DeleteData(ctx.Context, user, r); err != nil {
			return err
		}

		if err = client.DeleteBinaryData(user.Login, user.Password, r.Name); err != nil {
			return err
		}

		return nil
	}

}

func readDataFromFileArg(ctx *cli.Context, user *storage.User) (*storage.Record, error) {
	filename := args.GetFileArg(ctx)

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	encryptedData, err := encryptUserData(user, data)
	if err != nil {
		return nil, err
	}

	r := &storage.Record{Name: filename, Data: string(encryptedData), Revision: 1}

	return r, nil
}

func decryptUserData(
	user *storage.User,
	data []byte,
) ([]byte, error) {
	crypto, err := gophcrypto.New(user.CryptoKey)
	if err != nil {
		return nil, err
	}

	return crypto.Decrypt(data)
}

func encryptUserData(
	user *storage.User,
	data []byte,
) ([]byte, error) {
	crypto, err := gophcrypto.New([]byte(user.CryptoKey))
	if err != nil {
		return nil, err
	}

	return []byte(crypto.Encrypt(data)), nil
}
