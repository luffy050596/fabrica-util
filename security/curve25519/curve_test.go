package curve25519

import (
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/go-pantheon/fabrica-util/security/aes"
	"github.com/stretchr/testify/assert"
)

func TestKeyExchange(t *testing.T) {
	t.Parallel()

	serverPrivate, serverPublic, err := GenerateKeyPair()
	assert.NoError(t, err)

	clientPrivate, clientPublic, err := GenerateKeyPair()
	assert.NoError(t, err)

	serverSecret, err := ComputeSharedSecret(serverPrivate, clientPublic)
	assert.NoError(t, err)

	clientSecret, err := ComputeSharedSecret(clientPrivate, serverPublic)
	assert.NoError(t, err)

	assert.Equal(t, serverSecret, clientSecret)
}

func TestInvalidPublicKey(t *testing.T) {
	t.Parallel()

	invalidKey := make([]byte, 31)
	_, err := rand.Read(invalidKey)
	assert.NoError(t, err)

	_, err = ParsePublicKey(invalidKey)
	assert.Error(t, err)
}

func TestAllZeroPublicKey(t *testing.T) {
	t.Parallel()

	// Generate a valid private key
	privateKey, _, err := GenerateKeyPair()
	assert.NoError(t, err)

	// Create an all-zero public key (invalid)
	var zeroPublicKey [32]byte

	// This should fail because all-zero public key is a low order point
	_, err = ComputeSharedSecret(privateKey, zeroPublicKey)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "low order point")
}

func TestSharedSecret(t *testing.T) {
	t.Parallel()

	for i := range 100 {
		t.Run(fmt.Sprintf("test %d", i), func(t *testing.T) {
			t.Parallel()

			serverPrivate, _, err := GenerateKeyPair()
			assert.NoError(t, err)

			_, clientPublic, err := GenerateKeyPair()
			assert.NoError(t, err)

			secret, err := ComputeSharedSecret(serverPrivate, clientPublic)
			assert.NoError(t, err)

			cipher, err := aes.NewAESCipher(secret)
			assert.NoError(t, err)

			plaintext := []byte("Hello, world!")

			encrypted, err := cipher.Encrypt(plaintext)
			assert.NoError(t, err)

			decrypted, err := cipher.Decrypt(encrypted)
			assert.NoError(t, err)
			assert.Equal(t, plaintext, decrypted)
		})
	}
}

func BenchmarkKeyGeneration(b *testing.B) {
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _, err := GenerateKeyPair()
			if err != nil {
				b.Fatalf("key generation failed: %v", err)
			}
		}
	})
}

func BenchmarkSharedSecretComputation(b *testing.B) {
	// pre-generate fixed key pair
	serverPriv, serverPub, _ := GenerateKeyPair()
	clientPriv, clientPub, _ := GenerateKeyPair()

	// reset timer
	b.ResetTimer()
	b.ReportAllocs()

	// parallel test
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// test two direction key computation
			if _, err := ComputeSharedSecret(serverPriv, clientPub); err != nil {
				b.Fatal(err)
			}

			if _, err := ComputeSharedSecret(clientPriv, serverPub); err != nil {
				b.Fatal(err)
			}
		}
	})
}

var sink any // prevent compiler optimize

func BenchmarkKeyGenerationAndExchange(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// generate server key
		serverPriv, _, err := GenerateKeyPair()
		if err != nil {
			b.Fatal(err)
		}

		// generate client key
		_, clientPub, err := GenerateKeyPair()
		if err != nil {
			b.Fatal(err)
		}

		// compute shared secret
		secret, err := ComputeSharedSecret(serverPriv, clientPub)
		if err != nil {
			b.Fatal(err)
		}

		// prevent compiler optimize
		sink = secret
	}
}
