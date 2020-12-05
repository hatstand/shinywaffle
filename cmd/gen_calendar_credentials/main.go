package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var clientID = flag.String("client-id", "551173507406-3c50miifu3mtko6s90qt246ok61mv0b5.apps.googleusercontent.com", "Google OAuth Client ID")
var clientSecret = flag.String("client-secret", "", "Google OAuth Client Secret")

func main() {
	flag.Parse()

	conf := oauth2.Config{
		ClientID:     *clientID,
		ClientSecret: *clientSecret,
		Scopes: []string{
			"https://www.googleapis.com/auth/calendar.readonly",
		},
		Endpoint:    google.Endpoint,
		RedirectURL: "urn:ietf:wg:oauth:2.0:oob",
	}
	url := conf.AuthCodeURL("foo")
	fmt.Println(url)

	reader := bufio.NewReader(os.Stdin)
	code, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Failed to read auth code: %v", err)
	}

	tok, err := conf.Exchange(oauth2.NoContext, strings.TrimSpace(code))
	if err != nil {
		log.Fatalf("Failed to exchange token: %v", err)
	}
	creds, err := json.Marshal(tok)
	if err != nil {
		log.Fatalf("Failed to marshal token: %v", err)
	}
	fmt.Println(string(creds))
}
