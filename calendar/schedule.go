package calendar

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/patrickmn/go-cache"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) (*http.Client, error) {
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok, err = tokenFromEnv()
		if err != nil {
			tok, err = getTokenFromWeb(config)
			if err != nil {
				return nil, fmt.Errorf("Failed to get token from web: %v", err)
			}
			if err = saveToken(tokFile, tok); err != nil {
				return nil, fmt.Errorf("Failed to save token: %v", err)
			}
		}
	}
	return config.Client(context.Background(), tok), nil
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, fmt.Errorf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve token from web: %v", err)
	}
	return tok, nil
}

func tokenFromEnv() (*oauth2.Token, error) {
	tok := oauth2.Token{}
	if err := json.Unmarshal([]byte(os.Getenv("CALENDAR_TOKENS")), &tok); err != nil {
		return nil, err
	}
	return &tok, nil
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) error {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(token)
}

type CalendarScheduleService struct {
	cache   *cache.Cache
	service *calendar.Service
}

func NewCalendarScheduleService() (*CalendarScheduleService, error) {
	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON([]byte(os.Getenv("GOOGLE_CREDENTIALS")), calendar.CalendarReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse client secret file to config: %v", err)
	}
	client, err := getClient(config)
	if err != nil {
		return nil, fmt.Errorf("Failed to get calendar client: %v", err)
	}

	srv, err := calendar.New(client)
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve Calendar client: %v", err)
	}
	return &CalendarScheduleService{
		cache:   cache.New(time.Minute*10, time.Minute*20),
		service: srv,
	}, nil
}

func (srv *CalendarScheduleService) GetSchedule(calendarId string) ([]*calendar.TimePeriod, error) {
	cached, found := srv.cache.Get(calendarId)
	if found {
		return cached.([]*calendar.TimePeriod), nil
	}

	request := &calendar.FreeBusyRequest{
		Items: []*calendar.FreeBusyRequestItem{
			&calendar.FreeBusyRequestItem{
				Id: calendarId,
			},
		},
		TimeMin: time.Now().Add(time.Hour * -1).Format(time.RFC3339),
		TimeMax: time.Now().Add(time.Hour * 24 * 7).Format(time.RFC3339),
	}
	log.Printf("Fetching calendar for %s", calendarId)
	resp, err := srv.service.Freebusy.Query(request).Do()
	if err != nil {
		return nil, fmt.Errorf("Failed to request free/busy: %v", err)
	}
	srv.cache.Set(calendarId, resp.Calendars[calendarId].Busy, cache.DefaultExpiration)
	return resp.Calendars[calendarId].Busy, nil
}
