package gophcrypto

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
)

type Cryptographer struct {
	cipher cipher.AEAD
}

func New(key []byte) (*Cryptographer, error) {
	aesblock, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return nil, err
	}

	return &Cryptographer{cipher: aesgcm}, nil
}

func (c *Cryptographer) VecLen() int {
	return c.cipher.NonceSize()
}

func (c *Cryptographer) Encrypt(data []byte, vec []byte) string {
	dst := c.cipher.Seal(nil, vec, data, nil)

	return base64.RawStdEncoding.EncodeToString(dst)
}

func (c *Cryptographer) Decrypt(base64data []byte, vec []byte) ([]byte, error) {
	data := make([]byte, base64.RawStdEncoding.DecodedLen(len(base64data)))

	_, err := base64.RawStdEncoding.Decode(data, base64data)
	if err != nil {
		return nil, fmt.Errorf("base64 decode err=%w", err)
	}

	dst, err := c.cipher.Open(nil, vec, data, nil)
	if err != nil {
		return nil, fmt.Errorf("decode err=%w", err)
	}

	return dst, nil
}
