// Package aes provides AES encryption/decryption utilities with PKCS7 padding
package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"

	"github.com/go-pantheon/fabrica-util/errors"
)

// Cipher represents an AES cipher with a key and block
type Cipher struct {
	key   []byte
	block cipher.AEAD
}

// NewAESCipher creates a new AESCipher with the given key
func NewAESCipher(key []byte) (*Cipher, error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, errors.New("invalid key size: must be 16, 24, or 32 bytes")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create AES cipher")
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create AES GCM")
	}

	return &Cipher{
		key:   key,
		block: aead,
	}, nil
}

// EncryptAllowEmpty encrypts plaintext using AES-GCM, allowing empty data
func (c *Cipher) EncryptAllowEmpty(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}

	return c.Encrypt(data)
}

// Encrypt encrypts plaintext using AES-GCM
func (c *Cipher) Encrypt(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("data is empty")
	}

	nonce := make([]byte, c.block.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, errors.Wrap(err, "failed to generate nonce")
	}

	ciphertext := c.block.Seal(nonce, nonce, data, nil)

	return ciphertext, nil
}

// DecryptAllowEmpty decrypts ciphertext using AES-GCM, allowing empty data
func (c *Cipher) DecryptAllowEmpty(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}

	return c.Decrypt(data)
}

// Decrypt decrypts ciphertext using AES-GCM
func (c *Cipher) Decrypt(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("data is empty")
	}

	if len(data) < c.block.NonceSize() {
		return data, errors.New("cipher text is too short")
	}

	nonce := data[:c.block.NonceSize()]

	plaintext, err := c.block.Open(nil, nonce, data[c.block.NonceSize():], nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decrypt data")
	}

	return plaintext, nil
}
