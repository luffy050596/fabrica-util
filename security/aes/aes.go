// Package aes provides AES encryption/decryption utilities with PKCS7 padding
package aes

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"fmt"

	"github.com/pkg/errors"
)

// package error definitions
var (
	ErrCipherTextTooShort = errors.New("cipher text is too short")
	ErrInvalidPadding     = errors.New("invalid PKCS7 padding")
)

// MustEncrypt encrypts data or panics if an error occurs
func MustEncrypt(data []byte, key []byte) []byte {
	out, err := Encrypt(data, key)
	if err != nil {
		panic(fmt.Sprintf("aes encrypt error: %v", err))
	}

	return out
}

// Encrypt encrypts plaintext using AES-CBC with PKCS7 padding
func Encrypt(org []byte, key []byte) ([]byte, error) {
	if len(key) == 0 {
		return nil, errors.New("encryption key cannot be empty")
	}

	if len(org) == 0 {
		return nil, errors.New("plaintext cannot be empty")
	}

	// padding
	padded := pkcs7Padding(org, aes.BlockSize)

	// create cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create AES cipher")
	}

	ser := make([]byte, aes.BlockSize+len(padded))
	iv := ser[:aes.BlockSize]
	copy(iv, key[:aes.BlockSize])

	// encrypt
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ser[aes.BlockSize:], padded)

	return ser, nil
}

// Decrypt decrypts ciphertext using AES-CBC with PKCS7 padding
func Decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	if len(ciphertext) < aes.BlockSize {
		return nil, ErrCipherTextTooShort
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create AES cipher")
	}

	iv := ciphertext[:aes.BlockSize]
	org := make([]byte, len(ciphertext)-aes.BlockSize)

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(org, ciphertext[aes.BlockSize:])

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
