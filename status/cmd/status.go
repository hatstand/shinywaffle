package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"
	"os/signal"
	"sort"
	"sync"
	"time"

	"github.com/disintegration/imaging"
	"github.com/hatstand/shinywaffle/wirelesstag"
	"github.com/pbnjay/pixfont"
	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"

	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/spi/spireg"
	"periph.io/x/periph/experimental/devices/inky"
	"periph.io/x/periph/host"
)

var path = flag.String("template", "template.png", "Path to template image")

func drawLabel(m draw.Image, data string, x, y int) {
	pixfont.DrawString(m, x, y, data, color.White)
}

func drawTime(m draw.Image) {
	t := time.Now().Format("15:04:05 02/01")
	drawLabel(m, fmt.Sprintf("Updated: %s", t), 0, 104 - 8)
}

func drawWeather(m draw.Image) {
	iconsFile, err := zip.OpenReader("weather-icons-master.zip")
	if err != nil {
		log.Fatalf("Failed to open icons zip: %v", err)
	}
	for _, f := range iconsFile.File {
		if f.FileHeader.Name == "weather-icons-master/svg/wi-cloud.svg" {
			rc, err := f.Open()
			if err != nil {
				log.Fatalf("Failed to read wi-cloud.svg: %v", err)
			}
			defer rc.Close()
			icon, err := oksvg.ReadIconStream(rc)
			if err != nil {
				log.Fatalf("Failed to read SVG: %v", err)
			}

			w, h := int(icon.ViewBox.W), int(icon.ViewBox.H)
			img := image.NewRGBA(image.Rect(0, 0, w, h))
			scanner := rasterx.NewScannerGV(w, h, img, img.Bounds())
			raster := rasterx.NewDasher(w, h, scanner)
			icon.Draw(raster, 1.0)
			draw.Draw(imaging.Invert(m), image.Rect(212 - w, 0, 212, h), img, image.ZP, draw.Over)
			return
		}
	}
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
				sort.Slice(tags, func(i, j int) bool { return tags[i].Name < tags[j].Name })
				for i, v := range tags {
					s := fmt.Sprintf("%s: %.1f°C", v.Name, v.Temperature)
					log.Println(s)
					drawLabel(m, s, 0, 16*(i+1))
				}
				drawTime(m)
				drawWeather(m)

				dev.Draw(m.Bounds(), m, image.Point{0, 0})
			}()
		case <-c:
			return
		}
	}
}
