package main

import (
	"bufio"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/hatstand/shinywaffle/metar"
	"github.com/wcharczuk/go-chart"
)

const (
	baseURL = "https://www.ogimet.com/display_metars2.php"
	ICAO    = "EGLC"
)

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

	resp, err := http.Get(url.String())
	if err != nil {
		log.Fatalf("Failed to fetch METAR data: %v", err)
	}
	defer resp.Body.Close()

	doc, _ := goquery.NewDocumentFromReader(resp.Body)
	raw := doc.Find("pre").Text()
	METARs, err := metar.ParseMETARs(raw)
	if err != nil {
		log.Fatalf("Failed to parse METARs: %v", err)
	}

	var x []time.Time
	var y []float64
	for _, m := range METARs {
		x = append(x, m.DateTime)
		y = append(y, float64(m.Temperature))
	}

	graph := chart.Chart{
		XAxis: chart.XAxis{
			Style: chart.Style{
				Show: true,
			},
		},
		YAxis: chart.YAxis{
			Style: chart.Style{
				Show: true,
			},
		},
		Series: []chart.Series{
			chart.TimeSeries{
				XValues: x,
				YValues: y,
			},
		},
	}

	w := bufio.NewWriter(os.Stdout)
	err = graph.Render(chart.PNG, w)
}
