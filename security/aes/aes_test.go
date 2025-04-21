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
		name    string
		input   []byte
		wantErr bool
	}{
		{
			name:    "normal ascii text",
			input:   org,
			wantErr: false,
		},
		{
			name:    "empty input",
			input:   empty,
			wantErr: true,
		},
		{
			name:    "chinese characters",
			input:   utf8,
			wantErr: false,
		},
		{
			name:    "special characters",
			input:   special,
			wantErr: false,
		},
	}

	data, err := xrand.RandAlphaNumString(32)
	assert.Nil(t, err)

	aesKey := []byte(data)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Encrypt
			encrypted, err := Encrypt(tt.input, aesKey)
			if tt.wantErr {
				assert.NotNil(t, err)
				return
			}

			assert.Nil(t, err)

			// Decrypt
			decrypted, err := Decrypt(encrypted, aesKey)
			assert.Nil(t, err)
			assert.Equal(t, tt.input, decrypted)
		})
	}
}

func TestInvalidInputs(t *testing.T) {
	t.Parallel()

	// Test with nil inputs
	validKey, _ := xrand.RandAlphaNumString(32)

	_, err := Encrypt(nil, []byte(validKey))
	assert.NotNil(t, err)

	_, err = Encrypt(org, nil)
	assert.NotNil(t, err)

	_, err = Encrypt(empty, []byte(validKey))
	assert.NotNil(t, err)
}

func BenchmarkAESCBCEncrypt(b *testing.B) {
	data, _ := xrand.RandAlphaNumString(32)
	key := []byte(data)

	for i := 0; i < b.N; i++ {
		if _, err := Encrypt(org, key); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAESCBCDecrypt(b *testing.B) {
	data, _ := xrand.RandAlphaNumString(32)
	key := []byte(data)
	ser, _ := Encrypt(org, key)

	for i := 0; i < b.N; i++ {
		if _, err := Decrypt(ser, key); err != nil {
			b.Fatal(err)
		}
	}
}

func TestMustEncrypt(t *testing.T) {
	t.Parallel()

	data, err := xrand.RandAlphaNumString(32)
	require.Nil(t, err)

	key := []byte(data)

	// Test normal operation
	encrypted := MustEncrypt(org, key)
	decrypted, err := Decrypt(encrypted, key)
	assert.Nil(t, err)
	assert.Equal(t, org, decrypted)

	// Test panic with empty input
	assert.Panics(t, func() {
		MustEncrypt(empty, key)
	})
}

func TestGenerateAESKey(t *testing.T) {
	t.Parallel()

	data, err := xrand.RandAlphaNumString(32)
	require.Nil(t, err)
	fmt.Println(data)
}
