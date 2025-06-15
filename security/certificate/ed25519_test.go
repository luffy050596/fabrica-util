package certificate

import (
	"crypto/x509/pkix"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenKeyPair(t *testing.T) {
	t.Parallel()

	pair, err := GenKeyPair()
	require.NoError(t, err)

	// verify key pair
	err = ValidateKeyPair(pair)
	assert.NoError(t, err)
}

func TestSignAndVerify(t *testing.T) {
	t.Parallel()

	pair, err := GenKeyPair()
	require.NoError(t, err)

	// test message
	message := "Hello, ed25519 digital signature!"

	// sign message
	ret, err := SignMessage(pair.Pri, message)
	require.NoError(t, err)

	// verify signature
	valid := VerifySignResult(ret)
	assert.True(t, valid)

	// verify original message
	valid = VerifyMessage(pair.Pub, message, ret.Sign)
	assert.True(t, valid)
}

func TestCreateSelfSignedCertificate(t *testing.T) {
	t.Parallel()

	subject := pkix.Name{
		Country:            []string{"SG"},
		Province:           []string{"Singapore"},
		Locality:           []string{"Singapore"},
		Organization:       []string{"Go Pantheon"},
		OrganizationalUnit: []string{"Fabrica"},
		CommonName:         "janus.go-pantheon.dev",
	}

	// create self-signed certificate (valid for 365 days)
	cert, err := CreateSelfSignedCert(subject, 365)
	require.NoError(t, err)

	// verify certificate
	err = VerifyCert(cert.X509Cert)
	require.NoError(t, err)

	// verify import certificate
	cert2, err := ImportCertFromPEM(cert.CertPEM)
	require.NoError(t, err)

	assert.Equal(t, cert.X509Cert.PublicKey, cert2.PublicKey)
}

func TestPEMImportExport(t *testing.T) {
	t.Parallel()

	pair, err := GenKeyPair()
	require.NoError(t, err)

	// export private key to PEM
	priPEM, err := ExportPriToPEM(pair.Pri)
	require.NoError(t, err)

	// export public key to PEM
	pubPEM, err := ExportPubToPEM(pair.Pub)
	require.NoError(t, err)

	// import private key from PEM
	pri, err := ImportPriFromPEM(priPEM)
	require.NoError(t, err)
	assert.Equal(t, pair.Pri, pri)

	// import public key from PEM
	pub, err := ImportPubFromPEM(pubPEM)
	require.NoError(t, err)
	assert.Equal(t, pair.Pub, pub)

	// verify imported key pair
	importedPair := &KeyPair{
		Pri: pri,
		Pub: pub,
	}

	err = ValidateKeyPair(importedPair)
	require.NoError(t, err)
}

func TestBase64Encoding(t *testing.T) {
	t.Parallel()

	pair, err := GenKeyPair()
	require.NoError(t, err)

	// encode to base64
	priB64 := EncodeBase64(pair.Pri)
	pubB64 := EncodeBase64(pair.Pub)

	// decode base64
	pri, err := DecodeBase64(priB64)
	require.NoError(t, err)

	pub, err := DecodeBase64(pubB64)
	require.NoError(t, err)

	// verify decoded private key
	for i, b := range pair.Pri {
		assert.Equal(t, b, pri[i])
	}

	// verify decoded public key
	for i, b := range pair.Pub {
		assert.Equal(t, b, pub[i])
	}
}

func BenchmarkEd25519Tool_SignAndVerify(b *testing.B) {
	pair, err := GenKeyPair()
	require.NoError(b, err)

	msg := "Hello, ed25519 digital signature!"

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			result, err := SignMessage(pair.Pri, msg)
			require.NoError(b, err)
			assert.True(b, VerifySignResult(result))
		}
	})
}
