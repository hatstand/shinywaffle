package main

import (
	"flag"
	"log"
	"runtime"

	"github.com/davecgh/go-spew/spew"
	"github.com/hatstand/shinywaffle/wirelesstag"
)

var clientSecret = flag.String("secret", "", "OAuth2 client secret for WirelessTag")
var clientId = flag.String("client", "", "OAuth2 client id for WirelessTag")

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(12)

	if *clientSecret == "" || *clientId == "" {
		log.Fatalf("-secret and -client must be set")
	}

	v, err := wirelesstag.GetTags(*clientId, *clientSecret)
	if err != nil {
		log.Fatalf("Failed to fetch tags: %v", err)
	}
	spew.Dump(v)
}
