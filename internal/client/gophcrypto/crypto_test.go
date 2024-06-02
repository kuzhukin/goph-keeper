package gophcrypto

import (
	"crypto/aes"
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCrypto(t *testing.T) {

	data := []byte("12345")

	key, err := generateRandom(2 * aes.BlockSize)
	require.NoError(t, err)

	c, err := New(key)
	require.NoError(t, err)

	vec, err := generateRandom(c.VecLen())
	require.NoError(t, err)

	encrypted := c.Encrypt(data, vec)
	decrypted, err := c.Decrypt([]byte(encrypted), vec)
	require.NoError(t, err)

	fmt.Println("data", string(data))
	fmt.Println("decrypted", string(decrypted))

	require.Equal(t, data, decrypted)

}

func generateRandom(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}
