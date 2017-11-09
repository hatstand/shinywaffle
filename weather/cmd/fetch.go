package main

import (
	"flag"
	"log"

	"github.com/hatstand/shinywaffle/weather"
)

func main() {
	flag.Parse()

	w, err := weather.FetchCurrentWeather("London")
	if err != nil {
		log.Fatalf("Failed to fetch weather: %v", err)
	}
	log.Printf("%+v", w)
}
