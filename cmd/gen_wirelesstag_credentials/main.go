package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"golang.org/x/oauth2"
)

var clientID = flag.String("client-id", "44c1dfbd-85ed-4dd4-b4d3-51cccd3c067c", "WirelessTag OAuth2 client ID")
var clientSecret = flag.String("client-secret", "", "WirelessTag OAuth2 client secret")

func main() {
	flag.Parse()

	conf := oauth2.Config{
		ClientID:     *clientID,
		ClientSecret: *clientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://www.mytaglist.com/oauth2/authorize.aspx",
			TokenURL: "https://www.mytaglist.com/oauth2/access_token.aspx",
		},
		RedirectURL: "http://localhost:8080/",
	}
	url := conf.AuthCodeURL("foo")
	fmt.Println(url)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		log.Printf("Got code: %s", code)

		resp, err := http.Post(
			"https://www.mytaglist.com/oauth2/access_token.aspx",
			"application/x-www-form-urlencoded",
			strings.NewReader(fmt.Sprintf("client_id=%s&client_secret=%s&code=%s", conf.ClientID, conf.ClientSecret, code)),
		)
		if err != nil {
			log.Fatal(err)
		}
		d, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("Failed to read token body: %v", err)
		}
		fmt.Println(string(d))
		os.Exit(0)
	})
	if err := http.ListenAndServe("localhost:8080", nil); err != nil {
		log.Fatal(err)
	}
}
