package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	"github.com/hatstand/shinywaffle/metar"
	"github.com/hatstand/shinywaffle/wirelesstag"
	"github.com/wcharczuk/go-chart"
)

var clientSecret = flag.String("secret", "", "OAuth2 client secret for WirelessTag")
var clientId = flag.String("client", "", "OAuth2 client id for WirelessTag")
var icao = flag.String("icao", "EGLC", "ICAO code for an appropriately near airport")

type point struct {
	x time.Time
	y float64
}

func createSeries(data map[string]map[string][]float64, room string) *chart.TimeSeries {
	var series []point
	for date := range data {
		logs := data[date]
		d, _ := time.Parse("1/2/2006", date)
		for i, l := range logs[room] {
			if l == 0.0 {
				continue
			}
			t := time.Date(d.Year(), d.Month(), d.Day(), i, 0, 0, 0, time.UTC)
			series = append(series, point{
				x: t,
				y: l,
			})
		}
	}
	sort.Slice(series, func(i, j int) bool {
		return series[i].x.Before(series[j].x)
	})
	var tempX []time.Time
	var tempY []float64
	for _, t := range series {
		tempX = append(tempX, t.x)
		tempY = append(tempY, t.y)
	}
	return &chart.TimeSeries{
		Name:    room,
		XValues: tempX,
		YValues: tempY,
	}
}

func createSeriesFromMETAR(metar []*metar.METAR) *chart.TimeSeries {
	sort.Slice(metar, func(i, j int) bool {
		return metar[i].DateTime.Before(metar[j].DateTime)
	})
	var x []time.Time
	var y []float64
	for _, m := range metar {
		x = append(x, m.DateTime)
		y = append(y, float64(m.Temperature))
	}
	return &chart.TimeSeries{
		Name:    metar[0].ICAO,
		XValues: x,
		YValues: y,
	}
}

func main() {
	flag.Parse()

	data, err := wirelesstag.GetLogs(*clientId, *clientSecret)
	var dates []time.Time
	for k := range data {
		t, _ := time.Parse("1/2/2006", k)
		dates = append(dates, t)
	}
	sort.Slice(dates, func(i, j int) bool {
		return dates[i].Before(dates[j])
	})

	start := dates[0]
	finish := dates[len(dates)-1]

	log.Printf("Requesting METARs from %v to %v", start, finish)

	METARs, err := metar.FetchMETARs(start, finish, *icao)
	if err != nil {
		log.Fatalf("Failed to fetch METARs: %v", err)
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
			ValueFormatter: func(v interface{}) string {
				if vf, isFloat := v.(float64); isFloat {
					return fmt.Sprintf("%.1f°C", vf)
				}
				return ""
			},
			Range: &chart.ContinuousRange{
				Min: 0.0,
				Max: 36.0,
			},
		},
		Series: []chart.Series{
			createSeriesFromMETAR(METARs),
			createSeries(data, "Hall"),
			createSeries(data, "Bedroom"),
			createSeries(data, "Living Room"),
			createSeries(data, "Study"),
		},
	}

	graph.Elements = []chart.Renderable{
		chart.Legend(&graph),
	}

	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()
	graph.Render(chart.PNG, w)
}
