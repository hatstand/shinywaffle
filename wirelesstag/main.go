package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"runtime"

	"github.com/dchest/uniuri"
	"golang.org/x/oauth2"
)

var clientSecret = flag.String("secret", "OAuth2 client secret for WirelessTag", "")
var clientId = flag.String("client", "OAuth2 client id for WirelessTag", "")

func exchangeToken(code string) (*oauth2.Token, error) {
	response, err := http.PostForm("https://www.mytaglist.com/oauth2/access_token.aspx",
		url.Values{
			"client_id":     {*clientId},
			"client_secret": {*clientSecret},
			"code":          {code},
		})
	if err != nil {
		return nil, fmt.Errorf("Failed to exchange token: %v", err)
	}

	var token oauth2.Token
	err = json.NewDecoder(response.Body).Decode(&token)
	if err != nil {
		return nil, fmt.Errorf("Failed to decode response: %v", err)
	}
	return &token, nil
}

func main() {
	runtime.GOMAXPROCS(12)

	ctx := context.Background()
	conf := &oauth2.Config{
		ClientID:     *clientId,
		ClientSecret: *clientSecret,
		Scopes:       []string{},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://www.mytaglist.com/oauth2/authorize.aspx",
			TokenURL: "https://www.mytaglist.com/oauth2/access_token.aspx",
		},
		RedirectURL: "http://localhost:8080",
	}

	state := uniuri.New()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		log.Printf("Request: %v\n", r.URL)
		retState := r.URL.Query().Get("state")
		if retState != state {
			http.Error(w, "State inconsistent", 400)
			log.Fatalf("State does not match. Expected: %s Got: %s\n", state, retState)
		}
		code := r.URL.Query().Get("code")
		token, err := exchangeToken(code)
		if err != nil {
			http.Error(w, "Oops!", 400)
			log.Fatalf("Failed to exchange token: %v", err)
		}
		client := conf.Client(ctx, token)
		json := []byte(`{id:1, beepDuration:1}`)
		resp, err := client.Post("https://www.mytaglist.com/ethClient.asmx/Beep", "application/json", bytes.NewBuffer(json))
		if err != nil {
			log.Fatalf("Failed to fetch stuff: %v", err)
		}
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("Failed to read body: %v", err)
		}
		resp.Body.Close()
		fmt.Printf("%s\n", data)
	})

	url := conf.AuthCodeURL(state, oauth2.AccessTypeOffline)
	fmt.Printf("Visit the URL for the auth dialog: %v", url)

	http.ListenAndServe("localhost:8080", nil)
}
