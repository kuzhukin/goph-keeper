package action

import (
	"fmt"

	"github.com/kuzhukin/goph-keeper/internal/client/cli/args"
	"github.com/kuzhukin/goph-keeper/internal/client/storage"
	"github.com/kuzhukin/goph-keeper/internal/client/transport"
	"github.com/urfave/cli/v2"
)

func CreateCardActionHandler(
	user *storage.User,
	storage storage.WalletStorage,
	client *transport.Client,
) func(ctx *cli.Context) error {
	return func(ctx *cli.Context) error {
		card, err := args.GetBankCard(ctx)
		if err != nil {
			return fmt.Errorf("can't create card: %w", err)
		}

		data, err := storage.CreateCard(ctx.Context, user, card)
		if err != nil {
			return err
		}

		if err = client.CreateCardData(ctx.Context, user.Token, card.Number, data); err != nil {
			return err
		}

		return nil
	}
}

func DeleteCardActionHandler(
	user *storage.User,
	storage storage.WalletStorage,
	client *transport.Client,
) func(ctx *cli.Context) error {
	return func(ctx *cli.Context) error {
		number, err := args.GetCardNumber(ctx)
		if err != nil {
			return err
		}

		if err := storage.DeleteCard(ctx.Context, user, number); err != nil {
			return err
		}

		if err := client.DeleteCardData(ctx.Context, user.Token, number); err != nil {
			return err
		}

		return nil
	}
}

func ListCardActionHandler(
	user *storage.User,
	s storage.WalletStorage,
	client *transport.Client,
) func(ctx *cli.Context) error {
	return func(ctx *cli.Context) error {
		list, err := s.ListCard(ctx.Context, user)
		if err != nil {
			return err
		}

		for _, card := range list {
			fmt.Printf("holder: %s; num: %s; expiration: %v; cvv: %s\n", card.Owner, card.Number, card.ExpiryDate.Format(storage.ExpirationFormat), card.CvvCode)
		}

		return nil
	}
}
