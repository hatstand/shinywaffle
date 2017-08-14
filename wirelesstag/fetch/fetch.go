package main

import (
	"flag"
	"fmt"
	"log"
	"runtime"

	"github.com/hatstand/shinywaffle/wirelesstag"
)

var clientSecret = flag.String("secret", "", "OAuth2 client secret for WirelessTag")
var clientId = flag.String("client", "", "OAuth2 client id for WirelessTag")

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(12)

	tags, err := wirelesstag.GetTags(*clientId, *clientSecret)
	if err != nil {
		log.Fatalf("Failed to fetch tags: %v", err)
	}

	for _, t := range tags {
		fmt.Println(t)
	}
}
