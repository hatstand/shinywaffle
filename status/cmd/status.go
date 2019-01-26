package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/hatstand/shinywaffle/wirelesstag"
	"github.com/pbnjay/pixfont"

	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/spi/spireg"
	"periph.io/x/periph/experimental/devices/inky"
	"periph.io/x/periph/host"
)

var path = flag.String("template", "template.png", "Path to template image")

func drawLabel(m draw.Image, data string, x, y int) {
	pixfont.DrawString(m, x, y, data, color.White)
}

func main() {
	flag.Parse()

	state, err := host.Init()
	if err != nil {
		log.Fatalf("Failed to init periph: %v", err)
	}
	log.Printf("%+v", state)
	log.Printf("%+v", spireg.All())

	port, err := spireg.Open("")
	if err != nil {
		log.Fatalf("Failed to open SPI port: %v", err)
	}
	dc := gpioreg.ByName("22")
	reset := gpioreg.ByName("27")
	busy := gpioreg.ByName("17")

	dev, err := inky.New(port, dc, reset, busy, inky.Red)
	if err != nil {
		log.Fatalf("Failed to open inky: %v", err)
	}
	dev.SetBorder(inky.Black)

	file, err := os.Open(*path)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()
	img, err := png.Decode(file)
	if err != nil {
		log.Fatalf("Failed to decode %s as png: %v", *path, err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	var mu sync.Mutex

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			go func() {
				mu.Lock()
				defer mu.Unlock()
				b := img.Bounds()
				m := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
				draw.Draw(m, m.Bounds(), img, b.Min, draw.Src)

				tags, err := wirelesstag.GetTags("foo", "bar")
				if err != nil {
					log.Fatalf("Failed to fetch tags: %v", err)
				}
				for i, v := range tags {
					s := fmt.Sprintf("%s: %.1f°C", v.Name, v.Temperature)
					log.Println(s)
					drawLabel(m, s, 32, 8*(i+1))
				}

				dev.Draw(m.Bounds(), m, image.Point{0, 0})
			}()
		case <-c:
			return
		}
	}
}
