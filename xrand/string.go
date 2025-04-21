package xrand

import (
	"bytes"
	"math/rand/v2"

	"github.com/pkg/errors"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var charsetLen = len(charset)

// RandAlphaNumString generates a random alphanumeric string of the specified length
// It returns the generated string and any error encountered during generation
func RandAlphaNumString(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("length must be greater than 0")
	}

	r := randPool.Get().(*rand.Rand)
	defer randPool.Put(r)

	var buf bytes.Buffer

	buf.Grow(length)

	randomBytes := make([]byte, length)
	for range randomBytes {
		idx := r.IntN(charsetLen)
		buf.WriteByte(charset[idx])
	}

	return buf.String(), nil
}

// BytesN returns a random byte slice of length n
func BytesN(n int) []byte {
	if n <= 0 {
		return make([]byte, 0)
	}

	r := randPool.Get().(*rand.Rand)
	defer randPool.Put(r)

	bytes := make([]byte, n)

	if n >= 32 {
		chunks := n / 8
		remainder := n % 8

		for i := 0; i < chunks; i++ {
			value := r.Uint64()
			for j := 0; j < 8; j++ {
				bytes[i*8+j] = byte(value >> (j * 8))
			}
		}

		if remainder > 0 {
			value := r.Uint64()
			for j := 0; j < remainder; j++ {
				bytes[chunks*8+j] = byte(value >> (j * 8))
			}
		}
	} else {
		for i := range bytes {
			bytes[i] = byte(r.Uint32N(256))
		}
	}

	return bytes
}
