package ed25519

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
	PrivateKey ed25519.PrivateKey
	PublicKey  ed25519.PublicKey
}

// Certificate is X.509 certificate and related key pair
type Certificate struct {
	KeyPair  *KeyPair
	CertPEM  []byte
	CertDER  []byte
	X509Cert *x509.Certificate
}

// SignatureResult is signature result
type SignatureResult struct {
	Signature []byte
	Message   []byte
	PublicKey ed25519.PublicKey
}

// GenerateKeyPair generates new ed25519 key pair
func GenerateKeyPair() (*KeyPair, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ed25519 key pair: %w", err)
	}

	return &KeyPair{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	}, nil
}

// CreateSelfSignedCertificate creates self-signed certificate
func CreateSelfSignedCertificate(subject pkix.Name, validDays int) (*Certificate, error) {
	keyPair, err := GenerateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	// create certificate template
	template := x509.Certificate{
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
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, keyPair.PublicKey, keyPair.PrivateKey)
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

	return &Certificate{
		KeyPair:  keyPair,
		CertPEM:  certPEM,
		CertDER:  certDER,
		X509Cert: cert,
	}, nil
}

// SignMessage signs message with private key
func SignMessage(privateKey ed25519.PrivateKey, message string) (*SignatureResult, error) {
	return Sign(privateKey, []byte(message))
}

// Sign signs message with private key
func Sign(privateKey ed25519.PrivateKey, message []byte) (*SignatureResult, error) {
	if len(privateKey) != ed25519.PrivateKeySize {
		return nil, errors.New("invalid private key size")
	}

	signature := ed25519.Sign(privateKey, message)
	publicKey := privateKey.Public().(ed25519.PublicKey)

	return &SignatureResult{
		Signature: signature,
		Message:   message,
		PublicKey: publicKey,
	}, nil
}

// VerifyMessage verifies message signature
func VerifyMessage(publicKey ed25519.PublicKey, message string, signature []byte) bool {
	return Verify(publicKey, []byte(message), signature)
}

// VerifySignatureResult verifies SignatureResult
func VerifySignatureResult(result *SignatureResult) bool {
	return Verify(result.PublicKey, result.Message, result.Signature)
}

// Verify verifies signature
func Verify(publicKey ed25519.PublicKey, message, signature []byte) bool {
	if len(publicKey) != ed25519.PublicKeySize {
		return false
	}

	if len(signature) != ed25519.SignatureSize {
		return false
	}

	return ed25519.Verify(publicKey, message, signature)
}

// VerifyCertificate verifies certificate validity
func VerifyCertificate(cert *x509.Certificate) error {
	now := time.Now()

	if now.Before(cert.NotBefore) {
		return errors.New("certificate is not yet valid")
	}

	if now.After(cert.NotAfter) {
		return errors.New("certificate has expired")
	}

	return nil
}

// ExportPrivateKeyToPEM exports private key to PEM format
func ExportPrivateKeyToPEM(privateKey ed25519.PrivateKey) ([]byte, error) {
	privateKeyDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal private key: %w", err)
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyDER,
	})

	return privateKeyPEM, nil
}

// ExportPublicKeyToPEM exports public key to PEM format
func ExportPublicKeyToPEM(publicKey ed25519.PublicKey) ([]byte, error) {
	publicKeyDER, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %w", err)
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyDER,
	})

	return publicKeyPEM, nil
}

// ImportPrivateKeyFromPEM imports private key from PEM format
func ImportPrivateKeyFromPEM(pemData []byte) (ed25519.PrivateKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	ed25519PrivateKey, ok := privateKey.(ed25519.PrivateKey)
	if !ok {
		return nil, errors.New("not an ed25519 private key")
	}

	return ed25519PrivateKey, nil
}

// ImportPublicKeyFromPEM imports public key from PEM format
func ImportPublicKeyFromPEM(pemData []byte) (ed25519.PublicKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse public key")
	}

	ed25519PublicKey, ok := publicKey.(ed25519.PublicKey)
	if !ok {
		return nil, errors.New("not an ed25519 public key")
	}

	return ed25519PublicKey, nil
}

// ImportCertificateFromPEM imports
func ImportCertificateFromPEM(pemData []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return cert, nil
}

// EncodeBase64 encode bytes to base64 string
func EncodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// DecodeBase64 decode base64 string to bytes
func DecodeBase64(encoded string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(encoded)
}

// ValidateKeyPair check if the key pair is valid
func ValidateKeyPair(keyPair *KeyPair) error {
	if len(keyPair.PrivateKey) != ed25519.PrivateKeySize {
		return errors.Errorf("invalid private key size: expected %d, got %d", ed25519.PrivateKeySize, len(keyPair.PrivateKey))
	}

	if len(keyPair.PublicKey) != ed25519.PublicKeySize {
		return errors.Errorf("invalid public key size: expected %d, got %d", ed25519.PublicKeySize, len(keyPair.PublicKey))
	}

	expectedPublicKey, ok := keyPair.PrivateKey.Public().(ed25519.PublicKey)
	if !ok {
		return errors.New("public key does not match private key")
	}

	for i, b := range keyPair.PublicKey {
		if b != expectedPublicKey[i] {
			return errors.New("public key does not match private key")
		}
	}

	return nil
}
