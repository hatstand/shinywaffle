package calendar

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) (*http.Client, error) {
	tok, err := tokenFromEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to get calendar token from env: %w", err)
	}
	return config.Client(context.Background(), tok), nil
}

func tokenFromEnv() (*oauth2.Token, error) {
	tok := oauth2.Token{}
	if err := json.Unmarshal([]byte(os.Getenv("CALENDAR_TOKENS")), &tok); err != nil {
		return nil, err
	}
	return &tok, nil
}

type CalendarScheduleService struct {
	cache   *cache.Cache
	service *calendar.Service
	logger  *zap.SugaredLogger
}

func NewCalendarScheduleService(logger *zap.SugaredLogger) (*CalendarScheduleService, error) {
	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON([]byte(os.Getenv("GOOGLE_CREDENTIALS")), calendar.CalendarReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %w", err)
	}
	client, err := getClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to get calendar client: %w", err)
	}

	srv, err := calendar.New(client)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Calendar client: %w", err)
	}
	return &CalendarScheduleService{
		cache:   cache.New(time.Minute*10, time.Minute*20),
		service: srv,
		logger:  logger,
	}, nil
}

func (srv *CalendarScheduleService) GetSchedule(calendarId string) ([]*calendar.TimePeriod, error) {
	cached, found := srv.cache.Get(calendarId)
	if found {
		return cached.([]*calendar.TimePeriod), nil
	}

	request := &calendar.FreeBusyRequest{
		Items: []*calendar.FreeBusyRequestItem{
			{
				Id: calendarId,
			},
		},
		TimeMin: time.Now().Add(time.Hour * -1).Format(time.RFC3339),
		TimeMax: time.Now().Add(time.Hour * 24 * 7).Format(time.RFC3339),
	}
	srv.logger.Infof("Fetching calendar for %s", calendarId)
	resp, err := srv.service.Freebusy.Query(request).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to request free/busy: %w", err)
	}
	srv.cache.Set(calendarId, resp.Calendars[calendarId].Busy, cache.DefaultExpiration)
	return resp.Calendars[calendarId].Busy, nil
}
