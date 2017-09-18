package metar

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
	MaxDays = 30
)

var (
	METARMatcher   = regexp.MustCompile("^[0-9]{12}.*")
	METARRegexp    = regexp.MustCompile("^([0-9]{12}) (METAR|METAR COR|TAF|TAF AMD|TAF COR) ([A-Z]{4}) [0-9]{6}Z.*")
	TempRegexp     = regexp.MustCompile("(M?[0-9]{2})/(M?[0-9]{2})")
	PressureRegexp = regexp.MustCompile("Q([0-9]{4})")
)

type Type int

const (
	Routine  Type = iota
	Forecast      = iota
)

type METAR struct {
	DateTime         time.Time
	ICAO             string
	ReportType       Type
	Temperature      int
	DewPoint         int
	PressureMillibar int
}

func ParseMETARs(data string) ([]*METAR, error) {
	var ret []*METAR
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
				parsed, err := parseMETAR(buffer.String())
				if err != nil {
					return nil, fmt.Errorf("Failed to parse METAR: %v", err)
				}
				ret = append(ret, parsed)
			}
			buffer.Reset()
			buffer.WriteString(strings.TrimSpace(line))
		} else {
			buffer.WriteString(" ")
			buffer.WriteString(strings.TrimSpace(line))
		}
	}

	if strings.TrimSpace(buffer.String()) != "" {
		parsed, err := parseMETAR(buffer.String())
		if err != nil {
			return nil, fmt.Errorf("Failed to parse METAR: %v", err)
		}
		ret = append(ret, parsed)
	}
	return ret, nil
}

func reallyFetchMETARs(start time.Time, end time.Time, icao string) ([]*METAR, error) {
	log.Printf("Fetching METARs from: %v to %v", start, end)
	v := url.Values{}
	v.Set("lang", "en")
	v.Set("lugar", icao) // Location
	v.Set("tipo", "SA")  // Only METARs, no TAFs.
	v.Set("nil", "NO")
	v.Set("fmt", "txt")
	v.Set("send", "send")

	v.Set("ano", strconv.Itoa(start.Year()))
	v.Set("mes", fmt.Sprintf("%02d", start.Month()))
	v.Set("day", fmt.Sprintf("%02d", start.Day()))
	v.Set("hora", fmt.Sprintf("%02d", start.Hour()))

	v.Set("anof", strconv.Itoa(end.Year()))
	v.Set("mesf", fmt.Sprintf("%02d", end.Month()))
	v.Set("dayf", fmt.Sprintf("%02d", end.Day()))
	v.Set("horaf", fmt.Sprintf("%02d", end.Hour()))
	v.Set("minf", "59")

	url, _ := url.Parse(baseURL)
	url.RawQuery = v.Encode()

	log.Printf("METAR URL: %s", url.String())

	resp, err := http.Get(url.String())
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch METARs: %v", err)
	}
	defer resp.Body.Close()

	doc, _ := goquery.NewDocumentFromReader(resp.Body)
	raw := doc.Find("pre").Text()
	METARs, err := ParseMETARs(raw)
	if err != nil {
		return nil, fmt.Errorf("Failed to decode METARs: %v", err)
	}
	return METARs, nil
}

type page struct {
	start time.Time
	end   time.Time
}

func getPage(start time.Time, end time.Time) page {
	if end.Sub(start) <= MaxDays*time.Hour*24 {
		return page{
			start: start,
			end:   end,
		}
	} else {
		return page{
			start: start,
			end:   start.Add(MaxDays * time.Hour * 24),
		}
	}
}

func FetchMETARs(start time.Time, end time.Time, icao string) ([]*METAR, error) {
	var pages []page
	nextPage := getPage(start, end)
	for {
		pages = append(pages, nextPage)
		if nextPage.end == end {
			break
		}
		nextPage = getPage(nextPage.end, end)
	}
	var METARs []*METAR
	for _, page := range pages {
		m, err := reallyFetchMETARs(page.start, page.end, icao)
		if err != nil {
			return nil, fmt.Errorf("Failed to fetch METARs: %v", err)
		}
		METARs = append(METARs, m...)
	}
	return METARs, nil
}

func parseReportType(t string) Type {
	if t == "METAR" || t == "METAR COR" {
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

	pressureParsed := PressureRegexp.FindStringSubmatch(m)
	pressure, _ := strconv.Atoi(pressureParsed[1])

	return &METAR{
		DateTime:         dateTime,
		ICAO:             parsed[3],
		ReportType:       parseReportType(parsed[2]),
		Temperature:      temp,
		DewPoint:         dew,
		PressureMillibar: pressure,
	}, nil
}
