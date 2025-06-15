package ed25519

import (
	"crypto/x509/pkix"
	"fmt"
	"log"
)

// example 1: basic sign and verify
func ExampleBasicSignAndVerify() {
	// generate key pair
	keyPair, err := GenerateKeyPair()
	if err != nil {
		log.Fatalf("failed to generate key pair: %+v", err)
	}

	// message to sign
	message := "this is a message to sign"

	// sign message
	signResult, err := SignMessage(keyPair.PrivateKey, message)
	if err != nil {
		log.Fatalf("failed to sign message: %+v", err)
	}

	// verify signature
	isValid := VerifySignatureResult(signResult)

	fmt.Printf("message: %s\n", message)
	fmt.Printf("signature verification result: %t\n", isValid)
	fmt.Printf("signature (base64): %s\n", EncodeBase64(signResult.Signature))
	fmt.Printf("public key (base64): %s\n", EncodeBase64(signResult.PublicKey))
}

// example 2: create and verify X.509 certificate
func ExampleCreateCertificate() {
	// define certificate subject
	subject := pkix.Name{
		Country:            []string{"SG"},
		Province:           []string{"Singapore"},
		Locality:           []string{"Singapore"},
		Organization:       []string{"Pantheon"},
		OrganizationalUnit: []string{"Janus"},
		CommonName:         "janus.go-pantheon.dev",
	}

	// create self-signed certificate (valid for 365 days)
	cert, err := CreateSelfSignedCertificate(subject, 365)
	if err != nil {
		log.Fatalf("failed to create certificate: %+v", err)
	}

	// verify certificate
	err = VerifyCertificate(cert.X509Cert)
	if err != nil {
		log.Fatalf("failed to verify certificate: %+v", err)
	}

	// print certificate information
	fmt.Println("certificate created successfully!")
	fmt.Printf("certificate subject: %s\n", cert.X509Cert.Subject.String())
	fmt.Printf("validity period: %s to %s\n", cert.X509Cert.NotBefore, cert.X509Cert.NotAfter)
	fmt.Printf("serial number: %s\n", cert.X509Cert.SerialNumber.String())

	// print certificate PEM
	fmt.Printf("\ncertificate PEM:\n%s\n", string(cert.CertPEM))
}

// example 3: key import and export
func ExampleKeyImportExport() {
	fmt.Println("=== key import and export example ===")

	// generate key pair
	keyPair, err := GenerateKeyPair()
	if err != nil {
		log.Fatalf("failed to generate key pair: %+v", err)
	}

	// export private key to PEM
	privateKeyPEM, err := ExportPrivateKeyToPEM(keyPair.PrivateKey)
	if err != nil {
		log.Fatalf("failed to export private key to PEM: %+v", err)
	}

	// export public key to PEM
	publicKeyPEM, err := ExportPublicKeyToPEM(keyPair.PublicKey)
	if err != nil {
		log.Fatalf("failed to export public key to PEM: %+v", err)
	}

	fmt.Printf("private key PEM:\n%s\n", string(privateKeyPEM))
	fmt.Printf("public key PEM:\n%s\n", string(publicKeyPEM))

	// import private key from PEM
	importedPrivateKey, err := ImportPrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		log.Fatalf("failed to import private key from PEM: %+v", err)
	}

	importedPublicKey, err := ImportPublicKeyFromPEM(publicKeyPEM)
	if err != nil {
		log.Fatalf("failed to import public key from PEM: %+v", err)
	}

	// verify imported key pair
	importedKeyPair := &KeyPair{
		PrivateKey: importedPrivateKey,
		PublicKey:  importedPublicKey,
	}

	err = ValidateKeyPair(importedKeyPair)
	if err != nil {
		log.Fatalf("failed to validate imported key pair: %+v", err)
	}

	fmt.Printf("key successfully imported and validated!\n")
}

// example 4: file signature
func ExampleFileSignature() {
	fmt.Println("\n=== file signature example ===")

	// generate key pair
	keyPair, err := GenerateKeyPair()
	if err != nil {
		log.Fatalf("failed to generate key pair: %+v", err)
	}

	// mock file content
	fileContent := []byte(`
this is an important configuration file content:
{
  "version": "1.0.0",
  "author": "ed25519-tool",
  "config": {
    "debug": false,
    "max_connections": 100
  }
}
`)

	// sign file content
	signResult, err := Sign(keyPair.PrivateKey, fileContent)
	if err != nil {
		log.Fatalf("failed to sign file content: %+v", err)
	}

	// verify file signature
	isValid := Verify(keyPair.PublicKey, fileContent, signResult.Signature)

	fmt.Printf("file size: %d bytes\n", len(fileContent))
	fmt.Printf("file signature: %s\n", EncodeBase64(signResult.Signature))
	fmt.Printf("signature verification: %t\n", isValid)

	// mock file tampering
	tampered := []byte("this is a tampered file content")
	isTamperedValid := Verify(keyPair.PublicKey, tampered, signResult.Signature)
	fmt.Printf("tampered file verification: %t (should be false)\n", isTamperedValid)
}

// example 5: multiple signatures
func ExampleMultipleSignatures() {
	fmt.Println("\n=== multiple signatures example ===")

	// create multiple signers
	signers := make([]*KeyPair, 3)

	for i := range len(signers) {
		keyPair, err := GenerateKeyPair()
		if err != nil {
			log.Fatalf("failed to generate signer %d key pair: %+v", i+1, err)
		}

		signers[i] = keyPair
	}

	// message to sign
	message := "this is a message to sign"

	// each signer signs the message
	signatures := make([][]byte, len(signers))

	for i, signer := range signers {
		result, err := SignMessage(signer.PrivateKey, message)
		if err != nil {
			log.Fatalf("failed to sign message by signer %d: %+v", i+1, err)
		}

		signatures[i] = result.Signature
		fmt.Printf("signer %d signed: %s\n", i+1, EncodeBase64(result.Signature))
	}

	// verify all signatures
	fmt.Printf("\nverification result:\n")

	allValid := true

	for i, signature := range signatures {
		isValid := VerifyMessage(signers[i].PublicKey, message, signature)
		if !isValid {
			allValid = false
			break
		}
	}

	fmt.Printf("all signatures verified: %t\n", allValid)
}
