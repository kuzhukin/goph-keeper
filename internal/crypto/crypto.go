package crypto

import (
	"crypto/aes"
	"crypto/cipher"
)

type Cryptographer struct {
	cipher cipher.Block
}

func New(key []byte) (*Cryptographer, error) {
	cipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	return &Cryptographer{cipher: cipher}, nil
}

func (c *Cryptographer) Encrypt(data []byte) []byte {
	dst := make([]byte, aes.BlockSize)

	c.cipher.Encrypt(dst, data)

	return dst
}

func (c *Cryptographer) Decrypt(data []byte) []byte {
	dst := make([]byte, aes.BlockSize)

	c.cipher.Decrypt(dst, data)

	return dst
}
