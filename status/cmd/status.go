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

	"github.com/hatstand/shinywaffle/wirelesstag"

	"golang.org/x/image/font"
	"golang.org/x/image/font/inconsolata"
	"golang.org/x/image/math/fixed"
)

var path = flag.String("template", "template.png", "Path to template image")

var drawFont = inconsolata.Regular8x16

func drawLabel(m draw.Image, data string, x, y int) {
	drawer := font.Drawer{
		Dst:  m,
		Src:  image.NewUniform(color.RGBA{255, 255, 255, 255}),
		Face: drawFont,
		Dot:  fixed.Point26_6{fixed.Int26_6(x * 64), fixed.Int26_6(y * 64)},
	}
	drawer.DrawString(data)
}

func main() {
	flag.Parse()

	log.Printf("Font: %+v", drawFont)

	file, err := os.Open(*path)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()
	img, err := png.Decode(file)
	if err != nil {
		log.Fatalf("Failed to decode %s as png: %v", *path, err)
	}

	b := img.Bounds()
	m := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(m, m.Bounds(), img, b.Min, draw.Src)

	out, err := os.Create("out.png")
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}

	tags, err := wirelesstag.GetTags("foo", "bar")
	if err != nil {
		log.Fatalf("Failed to fetch tags: %v", err)
	}
	for i, v := range tags {
		drawLabel(m, fmt.Sprintf("%s: %.1f°C", v.Name, v.Temperature), 32, drawFont.Height*(i+1))
	}

	if err := png.Encode(out, m); err != nil {
		out.Close()
		log.Fatalf("Failed to encode output: %v", err)
	}

	if err := out.Close(); err != nil {
		log.Fatalf("Failed to close output file: %v", err)
	}
}
