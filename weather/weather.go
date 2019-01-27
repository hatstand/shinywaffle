package weather

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	baseUrl = "https://api.openweathermap.org/data/2.5/weather"
)

var apiKey = flag.String("api", "527c980a2885ec3ffc429e55e69c460a", "")

type EpochTime time.Time

func (et *EpochTime) UnmarshalJSON(data []byte) error {
	q, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return err
	}
	*(*time.Time)(et) = time.Unix(q, 0)
	return nil
}

type Location struct {
	CountryCode string `json:"country"`
	Sunrise     EpochTime
	Sunset      EpochTime
}

type Observation struct {
	CurrentTemp float32
	MaxTemp     float32
	MinTemp     float32
	Icon        string
	Humidity    int32
	Location    Location
}

func kelvinToCelsius(k float32) float32 {
	return k - 273.15
}

func FetchCurrentWeather(loc string) (*Observation, error) {
	u, _ := url.Parse(baseUrl)
	q := u.Query()
	q.Set("q", "London")
	q.Set("appid", *apiKey)
	u.RawQuery = q.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch weather: %v", err)
	}
	defer resp.Body.Close()

	type Condition struct {
		Description string `json:"description"`
		Icon        string `json:"icon"`
	}

	type Weather struct {
		Temp     float32 `json:"temp"`
		Pressure int32   `json:"pressure"`
		Humidity int32   `json:"humidity"`
		Min      float32 `json:"temp_min"`
		Max      float32 `json:"temp_max"`
	}

	type Message struct {
		C []Condition `json:"weather"`
		W Weather     `json:"main"`
		Location    Location `json:"sys"`
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read body: %v", err)
	}

	var m Message
	err = json.Unmarshal(data, &m)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse json: %v", err)
	}

	log.Println(m)

	return &Observation{
		CurrentTemp: kelvinToCelsius(m.W.Temp),
		Humidity:    m.W.Humidity,
		MinTemp:     kelvinToCelsius(m.W.Min),
		MaxTemp:     kelvinToCelsius(m.W.Max),
		Icon:        fmt.Sprintf("https://openweathermap.org/img/w/%s.png", m.C[0].Icon),
		Location:    m.Location,
	}, nil
}
