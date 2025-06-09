// Package aes provides AES encryption/decryption utilities with PKCS7 padding
package aes

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"

	"github.com/pkg/errors"
)

// package error definitions
var (
	ErrCipherTextTooShort = errors.New("cipher text is too short")
	ErrInvalidPadding     = errors.New("invalid PKCS7 padding")
)

// AESCipher represents an AES cipher with a key and block
type AESCipher struct {
	key   []byte
	block cipher.Block
}

// NewAESCipher creates a new AESCipher with the given key
func NewAESCipher(key []byte) (*AESCipher, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create AES cipher")
	}

	return &AESCipher{
		key:   key,
		block: block,
	}, nil
}

// Encrypt encrypts plaintext using AES-CBC with PKCS7 padding
func (c *AESCipher) Encrypt(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("plaintext cannot be empty")
	}

	padded := pkcs7Padding(data, aes.BlockSize)

	ser := make([]byte, aes.BlockSize+len(padded))
	iv := ser[:aes.BlockSize]
	copy(iv, c.key[:aes.BlockSize])

	mode := cipher.NewCBCEncrypter(c.block, iv)
	mode.CryptBlocks(ser[aes.BlockSize:], padded)

	return ser, nil
}

// Decrypt decrypts ciphertext using AES-CBC with PKCS7 padding
func (c *AESCipher) Decrypt(data []byte) ([]byte, error) {
	if len(data) < aes.BlockSize {
		return nil, ErrCipherTextTooShort
	}

	iv := data[:aes.BlockSize]
	org := make([]byte, len(data)-aes.BlockSize)

	mode := cipher.NewCBCDecrypter(c.block, iv)
	mode.CryptBlocks(org, data[aes.BlockSize:])

	return pkcs7UnPadding(org)
}

// pkcs7Padding adds PKCS7 padding to the data
func pkcs7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	data = append(data, padText...)

	return data
}

// pkcs7UnPadding removes PKCS7 padding from the data
func pkcs7UnPadding(origData []byte) ([]byte, error) {
	length := len(origData)
	unPadding := int(origData[length-1])

	if unPadding <= 0 || unPadding > length {
		return nil, ErrInvalidPadding
	}

	return origData[:(length - unPadding)], nil
}
