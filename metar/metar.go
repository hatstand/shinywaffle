package metar

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
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
