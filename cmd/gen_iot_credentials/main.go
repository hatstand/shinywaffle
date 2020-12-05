package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/lestrrat/go-jwx/jwk"
)

var privateRSAKeyPath = flag.String("private-key", "", "Path to RSA private key PEM file")

func main() {
	flag.Parse()

	d, err := ioutil.ReadFile(*privateRSAKeyPath)
	if err != nil {
		log.Fatal(err)
	}

	pemBlock, _ := pem.Decode(d)
	if len(pemBlock.Bytes) == 0 || pemBlock.Type != "PRIVATE KEY" {
		log.Fatal("Failed to read PEM")
	}

	key, err := x509.ParsePKCS8PrivateKey(pemBlock.Bytes)
	if err != nil {
		log.Fatal(err)
	}
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		log.Fatal("Failed to extract RSA key")
	}
	jKey, err := jwk.New(rsaKey)
	if err != nil {
		log.Fatal(err)
	}
	out, err := json.Marshal(jKey)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(out))
}
