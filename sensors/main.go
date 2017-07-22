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
	"github.com/hatstand/shinywaffle/sensors/sht31"
	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/rpi"
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

type Reading struct {
	Temperature float32
	Humidity    float32
	Location    string
	Timestamp   time.Time
}

func UploadReading(reading *Reading, srv *sheets.Service) error {
	log.Printf("Uploading reading for: %s", reading.Location)
	values := &sheets.ValueRange{
		MajorDimension: "ROWS",
		Values:         [][]interface{}{[]interface{}{reading.Location, reading.Temperature, reading.Humidity, reading.Timestamp.Unix()}},
	}
	req := srv.Spreadsheets.Values.Append(*spreadsheetId, "A1", values)
	req.ValueInputOption("RAW")
	_, err := req.Do()
	if err != nil {
		return fmt.Errorf("Failed to upload to sheets: %v", err)
	}
	return nil
}

func main() {
	runtime.GOMAXPROCS(8)
	flag.Parse()

	err := embd.InitI2C()
	if err != nil {
		log.Fatalf("Failed to initialize I2C: %v", err)
	}

	bus := embd.NewI2CBus(1)
	defer bus.Close()

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

	ch := make(chan *Reading, 10)
	sht31 := sht31.NewSHT31(bus, sht31.DEFAULT_ADDR)
	go func() {
		c := time.Tick(30 * time.Second)
		for _ = range c {
			log.Printf("Reading local sensor...")
			temp, humidity, err := sht31.ReadTempAndHum()
			if err != nil {
				log.Printf("Failed to read local sensor: %v", err)
				continue
			}
			ch <- &Reading{
				Location:    "Hallway",
				Temperature: temp,
				Humidity:    humidity,
				Timestamp:   time.Now(),
			}
		}
	}()

	packetCh := make(chan []byte, 10)
	cc1101 := shinywaffle.NewCC1101(packetCh)
	defer cc1101.Close()
	cc1101.SetSyncWord(SyncWord)

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)

	log.Printf("Listening for packets...\n")

	t := time.Tick(10 * time.Second)

	for {
		select {
		case p := <-packetCh:
			log.Printf("Received packet: %v\n", p)
			temp, humidity, err := parsePacket(p)
			if err != nil {
				log.Printf("Failed to parse packet: %v\n", err)
			} else {
				ch <- &Reading{
					Location:    "Kitchen",
					Temperature: temp,
					Humidity:    humidity,
					Timestamp:   time.Now(),
				}
			}
		case reading := <-ch:
			err = UploadReading(reading, srv)
			if err != nil {
				log.Printf("Failed to upload reading: %v", err)
			}
		case <-t:
			status, err := cc1101.Strobe(shinywaffle.SNOP)
			if err != nil {
				log.Println("Failed to read chip status")
			} else {
				log.Printf("Chip status: %#02x\n", status)
			}
		case <-signalCh:
			log.Printf("Shutting down...")
			return
		}
	}
}
