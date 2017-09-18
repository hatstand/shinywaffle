package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	baseURL = "https://www.ogimet.com/display_metars2.php"
	ICAO    = "EGLC"
)

var (
	METARMatcher = regexp.MustCompile("^[0-9]{12}.*")
	METARRegexp  = regexp.MustCompile("^([0-9]{12}) (METAR|TAF|TAF AMD|TAF COR) ([A-Z]{4}) [0-9]{6}Z.*")
	TempRegexp   = regexp.MustCompile("(M?[0-9]{2})/(M?[0-9]{2})")
)

type Type int

const (
	Routine  Type = iota
	Forecast      = iota
)

type METAR struct {
	DateTime    time.Time
	ICAO        string
	ReportType  Type
	Temperature int
	DewPoint    int
}

func parseMETARs(data string) []string {
	var ret []string
	lines := strings.Split(data, "\n")
	var buffer bytes.Buffer
	for _, line := range lines {
		// Skip comment lines.
		if strings.HasPrefix(line, "#") {
			continue
		}

		// Start of a METAR
		if METARMatcher.MatchString(line) {
			if buffer.Len() != 0 && strings.TrimSpace(buffer.String()) != "" {
				ret = append(ret, buffer.String())
			}
			buffer.Reset()
			buffer.WriteString(strings.TrimSpace(line))
		} else {
			buffer.WriteString(" ")
			buffer.WriteString(strings.TrimSpace(line))
		}
	}
	return ret
}

func parseReportType(t string) Type {
	if t == "METAR" {
		return Routine
	}
	return Forecast
}

func parseTemperature(t string) int {
	i, _ := strconv.Atoi(strings.TrimPrefix(t, "M"))
	if t[0] == 'M' {
		return i * -1
	}
	return i
}

func parseMETAR(m string) (*METAR, error) {
	parsed := METARRegexp.FindStringSubmatch(m)
	if parsed == nil {
		return nil, fmt.Errorf("Failed to parse METAR: %s", m)
	}

	dateTime, _ := time.Parse("200601021504", parsed[1])

	tempParsed := TempRegexp.FindStringSubmatch(m)
	temp := parseTemperature(tempParsed[1])
	dew := parseTemperature(tempParsed[2])

	return &METAR{
		DateTime:    dateTime,
		ICAO:        parsed[3],
		ReportType:  parseReportType(parsed[2]),
		Temperature: temp,
		DewPoint:    dew,
	}, nil
}

func main() {
	v := url.Values{}
	v.Set("lang", "en")
	v.Set("lugar", ICAO)
	v.Set("tipo", "SA")
	v.Set("ord", "REV")
	v.Set("nil", "NO")
	v.Set("fmt", "txt")
	v.Set("ano", "2017")
	v.Set("mes", "08")
	v.Set("day", "10")
	v.Set("hora", "00")
	v.Set("anof", "2017")
	v.Set("mesf", "08")
	v.Set("dayf", "31")
	v.Set("horaf", "23")
	v.Set("minf", "59")
	v.Set("send", "send")

	url, _ := url.Parse(baseURL)
	url.RawQuery = v.Encode()

	fmt.Printf("URL: %v\n", url)

	resp, err := http.Get(url.String())
	if err != nil {
		log.Fatalf("Failed to fetch METAR data: %v", err)
	}
	defer resp.Body.Close()

	doc, _ := goquery.NewDocumentFromReader(resp.Body)
	raw := doc.Find("pre").Text()
	METARs := parseMETARs(raw)
	for _, m := range METARs {
		p, err := parseMETAR(m)
		if err != nil {
			log.Fatalf("Parsing failed: %v", err)
		}
		fmt.Printf("Parsed: %v\n", p)
	}
}
