package main

import (
	"crypto/x509/pkix"
	"fmt"
	"log"

	"github.com/go-pantheon/fabrica-util/security/certificate"
)

func main() {
	pair, err := certificate.GenKeyPair()
	if err != nil {
		log.Fatal(err)
	}

	cert, err := certificate.CreateSelfSignedCert(pkix.Name{
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

	pri, err := certificate.ExportPriToPEM(pair.Pri)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\ncert PEM: \n%s\n", string(cert.CertPEM))
	fmt.Printf("\nprivate PEM: \n%s\n", string(pri))

	fmt.Printf("subject: %s\n", cert.X509Cert.Subject.String())
	fmt.Printf("issuer: %s\n", cert.X509Cert.Issuer.String())
	fmt.Printf("not before: %s\n", cert.X509Cert.NotBefore.String())
	fmt.Printf("not after: %s\n", cert.X509Cert.NotAfter.String())
	fmt.Printf("serial: %s\n", cert.X509Cert.SerialNumber.String())

	org := []byte("hello world")
	signRet, err := certificate.Sign(pair.Pri, org)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("signature: %s\n", certificate.EncodeBase64(signRet.Sign))

	valid := certificate.VerifySignResult(signRet)
	fmt.Printf("signature verification result: %t\n", valid)

	fmt.Println("succeed")
}
