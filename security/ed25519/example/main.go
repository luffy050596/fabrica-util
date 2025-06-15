package main

import (
	"crypto/x509/pkix"
	"fmt"
	"log"

	"github.com/go-pantheon/fabrica-util/security/ed25519"
)

func main() {
	keyPair, err := ed25519.GenerateKeyPair()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("private key: %s\n", ed25519.EncodeBase64(keyPair.PrivateKey))
	fmt.Printf("public key: %s\n", ed25519.EncodeBase64(keyPair.PublicKey))

	cert, err := ed25519.CreateSelfSignedCertificate(pkix.Name{
		CommonName: "janus.go-pantheon.dev",
		Country:    []string{"SG"},
		Province:   []string{"Singapore"},
		Locality:   []string{"Singapore"},
		Organization: []string{
			"Pantheon",
			"Janus",
		},
	}, 365)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\ncert PEM: \n%s\n", string(cert.CertPEM))
	fmt.Printf("cert DER: %s\n", ed25519.EncodeBase64(cert.CertDER))
	fmt.Printf("cert raw: %s\n", ed25519.EncodeBase64(cert.X509Cert.Raw))
	fmt.Printf("subject: %s\n", cert.X509Cert.Subject.String())
	fmt.Printf("issuer: %s\n", cert.X509Cert.Issuer.String())
	fmt.Printf("not before: %s\n", cert.X509Cert.NotBefore.String())
	fmt.Printf("not after: %s\n", cert.X509Cert.NotAfter.String())
	fmt.Printf("serial: %s\n", cert.X509Cert.SerialNumber.String())
}
