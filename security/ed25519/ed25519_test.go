package ed25519

import (
	"crypto/x509/pkix"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEd25519Tool_GenerateKeyPair(t *testing.T) {
	t.Parallel()

	keyPair, err := GenerateKeyPair()
	require.NoError(t, err)

	// verify key pair
	err = ValidateKeyPair(keyPair)
	assert.NoError(t, err)
}

func TestEd25519Tool_SignAndVerify(t *testing.T) {
	t.Parallel()

	keyPair, err := GenerateKeyPair()
	require.NoError(t, err)

	// test message
	message := "Hello, ed25519 digital signature!"

	// sign message
	result, err := SignMessage(keyPair.PrivateKey, message)
	require.NoError(t, err)

	// verify signature
	isValid := VerifySignatureResult(result)
	assert.True(t, isValid)

	// verify original message
	isValidMessage := VerifyMessage(keyPair.PublicKey, message, result.Signature)
	assert.True(t, isValidMessage)
}

func TestEd25519Tool_CreateSelfSignedCertificate(t *testing.T) {
	t.Parallel()

	subject := pkix.Name{
		Country:            []string{"CN"},
		Province:           []string{"Beijing"},
		Locality:           []string{"Beijing"},
		Organization:       []string{"Test Organization"},
		OrganizationalUnit: []string{"IT Department"},
		CommonName:         "test.example.com",
	}

	// create self-signed certificate (valid for 365 days)
	cert, err := CreateSelfSignedCertificate(subject, 365)
	require.NoError(t, err)

	// verify certificate
	err = VerifyCertificate(cert.X509Cert)
	require.NoError(t, err)
}

func TestEd25519Tool_PEMImportExport(t *testing.T) {
	t.Parallel()

	keyPair, err := GenerateKeyPair()
	require.NoError(t, err)

	// export private key to PEM
	privateKeyPEM, err := ExportPrivateKeyToPEM(keyPair.PrivateKey)
	require.NoError(t, err)

	// export public key to PEM
	publicKeyPEM, err := ExportPublicKeyToPEM(keyPair.PublicKey)
	require.NoError(t, err)

	// import private key from PEM
	importedPrivateKey, err := ImportPrivateKeyFromPEM(privateKeyPEM)
	require.NoError(t, err)

	// import public key from PEM
	importedPublicKey, err := ImportPublicKeyFromPEM(publicKeyPEM)
	require.NoError(t, err)

	// verify imported key pair
	importedKeyPair := &KeyPair{
		PrivateKey: importedPrivateKey,
		PublicKey:  importedPublicKey,
	}

	err = ValidateKeyPair(importedKeyPair)
	require.NoError(t, err)
}

func TestEd25519Tool_Base64Encoding(t *testing.T) {
	t.Parallel()

	keyPair, err := GenerateKeyPair()
	require.NoError(t, err)

	// encode to base64
	privateKeyB64 := EncodeBase64(keyPair.PrivateKey)
	publicKeyB64 := EncodeBase64(keyPair.PublicKey)

	// decode base64
	decodedPrivateKey, err := DecodeBase64(privateKeyB64)
	require.NoError(t, err)

	decodedPublicKey, err := DecodeBase64(publicKeyB64)
	require.NoError(t, err)

	// verify decoded private key
	for i, b := range keyPair.PrivateKey {
		assert.Equal(t, b, decodedPrivateKey[i])
	}

	// verify decoded public key
	for i, b := range keyPair.PublicKey {
		assert.Equal(t, b, decodedPublicKey[i])
	}
}

func BenchmarkEd25519Tool_SignAndVerify(b *testing.B) {
	keyPair, err := GenerateKeyPair()
	require.NoError(b, err)

	message := "Hello, ed25519 digital signature!"

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			result, err := SignMessage(keyPair.PrivateKey, message)
			require.NoError(b, err)
			assert.True(b, VerifySignatureResult(result))
		}
	})
}
