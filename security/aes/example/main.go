package main

import (
	"fmt"
	"log"

	"github.com/go-pantheon/fabrica-util/security/aes"
	"github.com/go-pantheon/fabrica-util/xrand"
)

func main() {
	key, err := xrand.RandAlphaNumString(32)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(key)
	fmt.Printf("key: %s\n", key)

	cipher, err := aes.NewAESCipher([]byte(key))
	if err != nil {
		log.Fatal(err)
	}

	org := []byte("hello world")
	ser, err := cipher.Encrypt(org)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("ser: %s\n", ser)

	dec, err := cipher.Decrypt(ser)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("dec: %s\n", dec)
	fmt.Println("succeed")
}
