package args

import (
	"errors"

	"github.com/kuzhukin/goph-keeper/internal/client/storage"
	"github.com/urfave/cli/v2"
)

func GetSecret(ctx *cli.Context) (*storage.Secret, error) {
	name := ctx.String("name")
	if len(name) == 0 {
		return nil, errors.New("bad secret's name")
	}

	key := ctx.String("key")
	if len(key) == 0 {
		return nil, errors.New("bad secret's key")
	}

	value := ctx.String("value")
	if len(value) == 0 {
		return nil, errors.New("bad secret's value")
	}

	return &storage.Secret{
		Name:  name,
		Key:   key,
		Value: value,
	}, nil
}
