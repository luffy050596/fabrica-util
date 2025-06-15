package certificate

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"

	"github.com/go-pantheon/fabrica-util/errors"
)

// KeyPair is ed25519 key pair
type KeyPair struct {
	Pri ed25519.PrivateKey
	Pub ed25519.PublicKey
}

// Cert is X.509 certificate and related key pair
type Cert struct {
	KeyPair  *KeyPair
	CertPEM  []byte
	CertDER  []byte
	X509Cert *x509.Certificate
}

// SignResult is signature result
type SignResult struct {
	Sign []byte
	Msg  []byte
	Pub  ed25519.PublicKey
}

// GenKeyPair generates new ed25519 key pair
func GenKeyPair() (*KeyPair, error) {
	pub, pri, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ed25519 key pair: %w", err)
	}

	return &KeyPair{
		Pri: pri,
		Pub: pub,
	}, nil
}

// CreateSelfSignedCert creates self-signed certificate
func CreateSelfSignedCert(subject pkix.Name, validDays int) (*Cert, error) {
	pair, err := GenKeyPair()
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	// create certificate tmpl
	tmpl := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               subject,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Duration(validDays) * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
	}

	// generate certificate DER encoding
	certDER, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, pair.Pub, pair.Pri)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	// parse certificate
	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	// convert to PEM format
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	return &Cert{
		KeyPair:  pair,
		CertPEM:  certPEM,
		CertDER:  certDER,
		X509Cert: cert,
	}, nil
}

// SignMessage signs message with private key
func SignMessage(pri ed25519.PrivateKey, msg string) (*SignResult, error) {
	return Sign(pri, []byte(msg))
}

// Sign signs message with private key
func Sign(pri ed25519.PrivateKey, msg []byte) (*SignResult, error) {
	if len(pri) != ed25519.PrivateKeySize {
		return nil, errors.New("invalid private key size")
	}

	sign := ed25519.Sign(pri, msg)
	pub := pri.Public().(ed25519.PublicKey)

	return &SignResult{
		Sign: sign,
		Msg:  msg,
		Pub:  pub,
	}, nil
}

// VerifyMessage verifies message signature
func VerifyMessage(pub ed25519.PublicKey, msg string, sign []byte) bool {
	return Verify(pub, []byte(msg), sign)
}

// VerifySignResult verifies SignatureResult
func VerifySignResult(ret *SignResult) bool {
	return Verify(ret.Pub, ret.Msg, ret.Sign)
}

// Verify verifies signature
func Verify(pub ed25519.PublicKey, msg, sign []byte) bool {
	if len(pub) != ed25519.PublicKeySize {
		return false
	}

	if len(sign) != ed25519.SignatureSize {
		return false
	}

	return ed25519.Verify(pub, msg, sign)
}

// VerifyCert verifies certificate validity
func VerifyCert(cert *x509.Certificate) error {
	now := time.Now()

	if now.Before(cert.NotBefore) {
		return errors.New("certificate is not yet valid")
	}

	if now.After(cert.NotAfter) {
		return errors.New("certificate has expired")
	}

	return nil
}

// ExportPriToPEM exports private key to PEM format
func ExportPriToPEM(pri ed25519.PrivateKey) ([]byte, error) {
	der, err := x509.MarshalPKCS8PrivateKey(pri)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal private key: %w", err)
	}

	pem := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: der,
	})

	return pem, nil
}

// ExportPubToPEM exports public key to PEM format
func ExportPubToPEM(pub ed25519.PublicKey) ([]byte, error) {
	der, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %w", err)
	}

	pem := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: der,
	})

	return pem, nil
}

// ExportPubFromCert extracts public key from x509.Certificate
func ExportPubFromCert(cert *x509.Certificate) (ed25519.PublicKey, error) {
	pub, ok := cert.PublicKey.(ed25519.PublicKey)
	if !ok {
		return nil, errors.Errorf("certificate does not contain an ed25519 public key, got type: %T", cert.PublicKey)
	}

	return pub, nil
}

// ImportPriFromPEM imports private key from PEM format
func ImportPriFromPEM(pemData []byte) (ed25519.PrivateKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("failed to decode private key PEM block")
	}

	pri, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	edpri, ok := pri.(ed25519.PrivateKey)
	if !ok {
		return nil, errors.New("not an ed25519 private key")
	}

	return edpri, nil
}

// ImportPubFromPEM imports public key from PEM format
func ImportPubFromPEM(pemData []byte) (ed25519.PublicKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("failed to decode public key PEM block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse public key")
	}

	edpub, ok := pub.(ed25519.PublicKey)
	if !ok {
		return nil, errors.New("not an ed25519 public key")
	}

	return edpub, nil
}

// ImportCertFromPEM imports
func ImportCertFromPEM(pemData []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("failed to decode certificate PEM block")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return cert, nil
}

// ValidateKeyPair check if the key pair is valid
func ValidateKeyPair(pair *KeyPair) error {
	if len(pair.Pri) != ed25519.PrivateKeySize {
		return errors.Errorf("invalid private key size: expected %d, got %d", ed25519.PrivateKeySize, len(pair.Pri))
	}

	if len(pair.Pub) != ed25519.PublicKeySize {
		return errors.Errorf("invalid public key size: expected %d, got %d", ed25519.PublicKeySize, len(pair.Pub))
	}

	pub, ok := pair.Pri.Public().(ed25519.PublicKey)
	if !ok {
		return errors.New("public key does not match private key")
	}

	for i, b := range pair.Pub {
		if b != pub[i] {
			return errors.New("public key does not match private key")
		}
	}

	return nil
}

// EncodeBase64 encode bytes to base64 string
func EncodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// DecodeBase64 decode base64 string to bytes
func DecodeBase64(encoded string) (ret []byte, err error) {
	if ret, err = base64.StdEncoding.DecodeString(encoded); err != nil {
		return nil, errors.Wrap(err, "failed to decode base64")
	}

	return ret, nil
}
