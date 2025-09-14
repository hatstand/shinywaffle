package main

import (
	"bufio"
	"flag"
	"log"
	"os"
	"time"

	"github.com/hatstand/shinywaffle/metar"
	chart "github.com/go-analyze/charts/chartdraw"
)

var icao = flag.String("icao", "EGLC", "ICAO code of an airport")

func main() {
	finish := time.Now().Round(time.Hour)
	start := finish.Add(-30 * 24 * time.Hour)

	METARs, err := metar.FetchMETARs(start, finish, *icao)
	if err != nil {
		log.Fatalf("Failed to fetch METARs: %v", err)
	}

	var x []time.Time
	var y []float64
	for _, m := range METARs {
		x = append(x, m.DateTime)
		y = append(y, float64(m.Temperature))
	}

	graph := chart.Chart{
		Series: []chart.Series{
			chart.TimeSeries{
				XValues: x,
				YValues: y,
			},
		},
	}

	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()
	err = graph.Render(chart.PNG, w)
}
