package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var clientID = flag.String("client-id", "551173507406-3c50miifu3mtko6s90qt246ok61mv0b5.apps.googleusercontent.com", "Google OAuth Client ID")
var clientSecret = flag.String("client-secret", "", "Google OAuth Client Secret")

func main() {
	flag.Parse()
	ctx := context.Background()

	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	fmt.Println("Listening on port", port)
	srv := &http.Server{}
	go func() {
		if err := srv.Serve(l); err != nil {
			log.Fatal(err)
		}
	}()

	conf := oauth2.Config{
		ClientID:     *clientID,
		ClientSecret: *clientSecret,
		Scopes: []string{
			"https://www.googleapis.com/auth/calendar.readonly",
		},
		Endpoint:    google.Endpoint,
		RedirectURL: fmt.Sprintf("http://localhost:%d", port),
	}
	url := conf.AuthCodeURL("foo")
	fmt.Println(url)

	reader := bufio.NewReader(os.Stdin)
	code, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Failed to read auth code: %v", err)
	}

	tok, err := conf.Exchange(ctx, strings.TrimSpace(code))
	if err != nil {
		log.Fatalf("Failed to exchange token: %v", err)
	}
	creds, err := json.Marshal(tok)
	if err != nil {
		log.Fatalf("Failed to marshal token: %v", err)
	}
	fmt.Println(string(creds))
}
