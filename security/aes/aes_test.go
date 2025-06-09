package aes

import (
	"fmt"
	"testing"

	"github.com/go-pantheon/fabrica-util/xrand"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	org     = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	empty   = []byte("")
	utf8    = []byte("测试中文加密解密")
	special = []byte("!@#$%^&*()_+-=[]{}|;:,.<>?")
)

func TestAESCBCCodec(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input []byte
	}{
		{
			name:  "normal ascii text",
			input: org,
		},
		{
			name:  "chinese characters",
			input: utf8,
		},
		{
			name:  "special characters",
			input: special,
		},
	}

	data, err := xrand.RandAlphaNumString(32)
	assert.Nil(t, err)

	aesKey := []byte(data)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Encrypt
			cipher, err := NewAESCipher(aesKey)
			require.Nil(t, err)
			encrypted, err := cipher.Encrypt(tt.input)
			assert.Nil(t, err)

			// Decrypt
			decrypted, err := cipher.Decrypt(encrypted)
			assert.Nil(t, err)
			assert.Equal(t, tt.input, decrypted)
		})
	}
}

func TestInvalidInputs(t *testing.T) {
	t.Parallel()

	key, _ := xrand.RandAlphaNumString(32)

	aes, err := NewAESCipher([]byte(key))
	require.Nil(t, err)

	_, err = aes.Encrypt(nil)
	require.Error(t, err)

	_, err = aes.Encrypt(empty)
	require.Error(t, err)

	_, err = aes.Decrypt(nil)
	require.Error(t, err)

	_, err = aes.Decrypt(empty)
	require.Error(t, err)
}

func BenchmarkAESCBCEncrypt(b *testing.B) {
	data, err := xrand.RandAlphaNumString(32)
	require.Nil(b, err)

	cipher, err := NewAESCipher([]byte(data))
	require.Nil(b, err)

	for i := 0; i < b.N; i++ {
		if _, err := cipher.Encrypt(org); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAESCBCDecrypt(b *testing.B) {
	data, err := xrand.RandAlphaNumString(32)
	require.Nil(b, err)

	cipher, err := NewAESCipher([]byte(data))
	require.Nil(b, err)

	ser, err := cipher.Encrypt(org)
	require.Nil(b, err)

	for i := 0; i < b.N; i++ {
		if _, err := cipher.Decrypt(ser); err != nil {
			b.Fatal(err)
		}
	}
}

func TestGenerateAESKey(t *testing.T) {
	t.Parallel()

	data, err := xrand.RandAlphaNumString(32)
	require.Nil(t, err)
	fmt.Println(data)
}
