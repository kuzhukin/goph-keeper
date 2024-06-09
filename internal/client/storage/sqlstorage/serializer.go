package sqlstorage

import (
	"encoding/base64"
	"encoding/json"

	"github.com/kuzhukin/goph-keeper/internal/client/gophcrypto"
	"github.com/kuzhukin/goph-keeper/internal/client/storage"
)

type DbSerializer struct {
	crypt *gophcrypto.Cryptographer
}

func NewSerializer(crypt *gophcrypto.Cryptographer) *DbSerializer {
	return &DbSerializer{crypt: crypt}
}

func (s *DbSerializer) SerializeBankCard(card *storage.BankCard) (string, error) {
	marshaled, err := json.Marshal(card)
	if err != nil {
		return "", err
	}

	s.crypt.Encrypt(marshaled)

	based64Data := base64.StdEncoding.EncodeToString(marshaled)

	return based64Data, nil
}

func (s *DbSerializer) DeserializeBankCard(base64data string) (*storage.BankCard, error) {
	encoded, err := base64.StdEncoding.Strict().DecodeString(base64data)
	if err != nil {
		return nil, err
	}

	data, err := s.crypt.Decrypt(encoded)
	if err != nil {
		return nil, err
	}

	card := &storage.BankCard{}
	err = json.Unmarshal(data, card)
	if err != nil {
		return nil, err
	}

	return card, nil
}
