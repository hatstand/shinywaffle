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
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/dchest/uniuri"
	"golang.org/x/oauth2"
)

var clientSecret = flag.String("secret", "", "OAuth2 client secret for WirelessTag")
var clientId = flag.String("client", "", "OAuth2 client id for WirelessTag")
var port = flag.Int("port", 8080, "Port for OAuth2 flow")

type Tag struct {
	Name             string  `json:"name"`
	Temperature      float64 `json:"temperature"`
	UUID             string  `json:"uuid"`
	SignaldBm        float64 `json:"signaldBm"`
	BatteryRemaining float64 `json:"batteryRemaining"`
	Humidity         float64 `json:"cap"`
}

type TagList struct {
	Tag []Tag `json:"d"`
}

func exchangeToken(config *oauth2.Config, code string) (*oauth2.Token, error) {
	response, err := http.PostForm(
		config.Endpoint.TokenURL,
		url.Values{
			"client_id":     {config.ClientID},
			"client_secret": {config.ClientSecret},
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

func tokenCacheFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("Failed to get cache file: %v", err)
	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir, url.QueryEscape("mytaglist.json")), nil
}

func saveToken(file string, token *oauth2.Token) {
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("Failed to open token cache file: %v", err)
	}
	defer f.Close()
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	return t, err
}

func tokenFromWeb(ctx context.Context, config *oauth2.Config) (token *oauth2.Token, err error) {
	state := uniuri.New()
	mux := http.NewServeMux()
	srv := &http.Server{
		Addr:    "localhost:" + strconv.Itoa(*port),
		Handler: mux,
	}
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
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
		token, err = exchangeToken(config, code)
		if err != nil {
			http.Error(w, "Oops!", 400)
			log.Fatalf("Failed to exchange token: %v", err)
		}
		stopCtx, cancel := context.WithDeadline(ctx, time.Now().Add(2*time.Second))
		defer cancel()
		srv.Shutdown(stopCtx)
	})

	url := config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	fmt.Printf("Visit the URL for the auth dialog: %v", url)
	srv.ListenAndServe()
	return token, nil
}

func getClient(ctx context.Context, config *oauth2.Config) *http.Client {
	cacheFile, err := tokenCacheFile()
	if err != nil {
		log.Fatalf("Unable to get token cache file: %v", err)
	}
	token, err := tokenFromFile(cacheFile)
	if err != nil {
		token, err = tokenFromWeb(ctx, config)
		if err != nil {
			log.Fatalf("Unable to get token from web: %v", err)
		}
		saveToken(cacheFile, token)
	}
	return config.Client(ctx, token)
}

func main() {
	flag.Parse()
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
		RedirectURL: "http://localhost:" + strconv.Itoa(*port) + "/",
	}

	client := getClient(ctx, conf)
	resp, err := client.Post("https://www.mytaglist.com/ethClient.asmx/GetTagListCached", "application/json", bytes.NewBuffer([]byte(`{}`)))
	if err != nil {
		log.Fatalf("Failed to fetch stuff: %v", err)
	}

	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	var tags TagList
	err = json.Unmarshal(data, &tags)
	if err != nil {
		log.Fatalf("Failed to decode JSON: %v", err)
	}

	for _, t := range tags.Tag {
		fmt.Println(t)
	}
}
