package rsa

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"testing"

	"github.com/go-pantheon/fabrica-util/xrand"
	"github.com/stretchr/testify/assert"
)

func TestRSAEncryptDecrypt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		keyBits   int
		plaintext []byte
	}{
		{
			name:      "short text with 2048 bits key",
			keyBits:   2048,
			plaintext: []byte("Hello, World!"),
		},
		{
			name:      "longer text with 4096 bits key",
			keyBits:   4096,
			plaintext: []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, pri, pubBytes, _, err := generateTestKeyPair(tt.keyBits)
			assert.NoError(t, err)

			pub, err := ParsePublicKey(pubBytes)
			assert.NoError(t, err)

			// Test encryption
			encrypted, err := Encrypt(pub, tt.plaintext)
			assert.NoError(t, err)
			assert.NotEmpty(t, encrypted)

			// Test decryption
			decrypted, err := Decrypt(pri, encrypted)
			assert.NoError(t, err)
			assert.Equal(t, tt.plaintext, decrypted)
		})
	}
}

func TestRSAKeyMarshaling(t *testing.T) {
	t.Parallel()
	// Generate test key pair
	_, pri, pubBytes, priBytes, err := generateTestKeyPair(4096)
	assert.NoError(t, err)

	// Test public key unmarshaling
	pub, err := x509.ParsePKIXPublicKey(pubBytes)
	assert.NoError(t, err)

	// Test private key unmarshaling
	pri2, err := x509.ParsePKCS8PrivateKey(priBytes)
	assert.NoError(t, err)

	// Test marshaling back
	pubBytes2, err := x509.MarshalPKIXPublicKey(pub)
	assert.NoError(t, err)
	priBytes2, err := x509.MarshalPKCS8PrivateKey(pri2)
	assert.NoError(t, err)

	// Verify marshaled bytes are identical
	assert.Equal(t, pubBytes, pubBytes2)
	assert.Equal(t, priBytes, priBytes2)

	// Verify keys functionality
	testData := []byte("test encryption after marshaling")
	encrypted, err := Encrypt(pub.(*rsa.PublicKey), testData)
	assert.NoError(t, err)

	decrypted, err := Decrypt(pri, encrypted)
	assert.NoError(t, err)
	assert.Equal(t, testData, decrypted)
}

func TestRSASignVerify(t *testing.T) {
	t.Parallel()

	_, pri, _, _, err := generateTestKeyPair(2048)
	assert.NoError(t, err)

	pub := &pri.PublicKey

	testData := []byte("data to sign")
	hashed := sha256.Sum256(testData)

	// Test signing
	signature, err := rsa.SignPKCS1v15(rand.Reader, pri, crypto.SHA256, hashed[:])
	assert.NoError(t, err)
	assert.NotEmpty(t, signature)

	// Test verification
	err = rsa.VerifyPKCS1v15(pub, crypto.SHA256, hashed[:], signature)
	assert.NoError(t, err)

	// Test verification with modified data
	hashedModified := sha256.Sum256([]byte("modified data"))
	err = rsa.VerifyPKCS1v15(pub, crypto.SHA256, hashedModified[:], signature)
	assert.Error(t, err)
}

func BenchmarkRSAEncrypt(b *testing.B) {
	pub, _, _, _, err := generateTestKeyPair(4096)
	assert.NoError(b, err)

	data, _ := xrand.RandAlphaNumString(256)
	org := []byte(data)

	for i := 0; i < b.N; i++ {
		_, _ = Encrypt(pub, org)
	}
}

func BenchmarkRSADecrypt(b *testing.B) {
	pub, pri, _, _, err := generateTestKeyPair(4096)
	assert.NoError(b, err)

	data, _ := xrand.RandAlphaNumString(30)
	org := []byte(data)

	dst, _ := rsa.EncryptPKCS1v15(rand.Reader, pub, org)
	for i := 0; i < b.N; i++ {
		_, _ = Decrypt(pri, dst)
	}
}

func BenchmarkGenRsaKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _, _, _, err := generateTestKeyPair(4096)
		assert.NoError(b, err)
	}
}

func generateTestKeyPair(bits int) (*rsa.PublicKey, *rsa.PrivateKey, []byte, []byte, error) {
	pri, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	pub := &pri.PublicKey
	pubBytes, err := x509.MarshalPKIXPublicKey(pub)

	if err != nil {
		return nil, nil, nil, nil, err
	}

	priBytes, err := x509.MarshalPKCS8PrivateKey(pri)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	return pub, pri, pubBytes, priBytes, nil
}

func TestGenerateKeyPair(t *testing.T) {
	t.Parallel()

	pub, pri, pubBytes, priBytes, err := generateTestKeyPair(4096)
	assert.NoError(t, err)
	assert.NotNil(t, pub)
	assert.NotNil(t, pri)
	assert.NotEmpty(t, pubBytes)
	assert.NotEmpty(t, priBytes)
	t.Logf("pub: %v", base64.URLEncoding.EncodeToString(pubBytes))
	t.Logf("pri: %v", base64.URLEncoding.EncodeToString(priBytes))
}
