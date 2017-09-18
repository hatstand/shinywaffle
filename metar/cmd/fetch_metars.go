package main

import (
	"bufio"
	"log"
	"os"
	"time"

	"github.com/hatstand/shinywaffle/metar"
	"github.com/wcharczuk/go-chart"
)

func main() {
	finish := time.Now().Round(time.Hour)
	start := finish.Add(-30 * 24 * time.Hour)

	METARs, err := metar.FetchMETARs(start, finish)
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
