package args

import (
	"errors"
	"fmt"

	"github.com/kuzhukin/goph-keeper/internal/client/storage"
	"github.com/urfave/cli/v2"
)

func GetBankCard(ctx *cli.Context) (*storage.BankCard, error) {
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
		return nil, fmt.Errorf("bad card's cvv=%s", ctx.String("cvv"))
	}

	owner, ok := validateCardOwner(ctx.String("owner"))
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

func GetCardNumber(ctx *cli.Context) (string, error) {
	number, ok := validateCardNumber(ctx.String("number"))
	if !ok {
		return "", errors.New("bad card number")
	}

	return number, nil
}
