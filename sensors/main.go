package main

import (
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"runtime"
	"time"

	"github.com/hatstand/shinywaffle"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
)

const (
	SyncWord = 0x4242
)

var spreadsheetId = flag.String("sheet", "", "Id of Google Sheet")

func getClient(ctx context.Context, config *oauth2.Config) *http.Client {
	cacheFile, err := tokenCacheFile()
	if err != nil {
		log.Fatalf("Unable to get path to cached credential file: %v", err)
	}
	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(cacheFile, tok)
	}
	return config.Client(ctx, tok)
}

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code: %v", code)
	}

	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

func tokenCacheFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir,
		url.QueryEscape("sheets.googleapis.com-go-quickstart.json")), nil
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	return t, err
}

func saveToken(file string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func parsePacket(packet []byte) (float32, float32, error) {
	if len(packet) != 4 {
		return 0.0, 0.0, fmt.Errorf("Expected packet of size: 4 but was: %d", len(packet))
	}
	temp := float32(binary.BigEndian.Uint16(packet[:2])) / 100.0
	humidity := float32(binary.BigEndian.Uint16(packet[2:])) / 100.0
	log.Printf("Temperature: %.1fC Humidity: %.1f%%", temp, humidity)
	return temp, humidity, nil
}

func main() {
	runtime.GOMAXPROCS(4)
	flag.Parse()

	ctx := context.Background()
	b, err := ioutil.ReadFile("client_secret.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(ctx, config)

	srv, err := sheets.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets Client: %v", err)
	}

	packetCh := make(chan []byte, 10)
	cc1101 := shinywaffle.NewCC1101(packetCh)
	defer cc1101.Close()
	cc1101.SetSyncWord(SyncWord)

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)

	log.Printf("Listening for packets...\n")

	for {
		select {
		case p := <-packetCh:
			log.Printf("Received packet: %v\n", p)
			temp, humidity, err := parsePacket(p)
			if err != nil {
				log.Printf("Failed to parse packet: %v\n", err)
			} else {
				values := &sheets.ValueRange{
					MajorDimension: "ROWS",
					Values:         [][]interface{}{[]interface{}{"Kitchen", temp, humidity, time.Now()}},
				}
				_, err := srv.Spreadsheets.Values.Append(*spreadsheetId, "A1", values).ValueInputOption("RAW").Do()
				if err != nil {
					log.Printf("Failed to update spreadsheet: %v", err)
				}
			}
		case <-signalCh:
			return
		}
	}
}
