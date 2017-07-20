package main

import (
	"flag"
	"log"
	"time"

	"github.com/hatstand/shinywaffle"
)

// Radiator Addresses.
const (
	Kitchen = 0x2b7e
	Study   = 0x2e04
	Living  = 0x2bdb
	Bedroom = 0x2b76
)

// Radiator Modes.
const (
	Day   = 0x05
	Night = 0x03
	Cold  = 0x09
	Off   = 0x60
	Auto  = 0x11
)

func convertTemp(temp float32) byte {
	// Temperatures are represented as bytes in half-degree intervals, i.e. 0.5C -> 1, 1C -> 2
	return byte(int(temp*10) * 2 / 10)
}

func NewPacket(address uint16, mode byte, dayTemp float32, nightTemp float32, coldTemp float32) []byte {
	packet := []byte{0x57, 0x16, 0x0a}
	packet = append(packet, byte((address&0xff00)>>8))
	packet = append(packet, byte(address&0x00ff))
	packet = append(packet, mode)
	packet = append(packet, convertTemp(dayTemp))
	packet = append(packet, convertTemp(nightTemp))
	packet = append(packet, convertTemp(coldTemp))
	return packet
}

func main() {
	flag.Parse()

	packetCh := make(chan []byte, 10)
	cc1101 := shinywaffle.NewCC1101(packetCh)
	defer cc1101.Close()
	cc1101.SetSyncWord(0xd391)

	time.Sleep(5 * time.Second)
	packet := NewPacket(Kitchen, Off, 25.0, 15.0, 10.0)
	cc1101.Send(packet)
	cc1101.Send(packet)
	cc1101.Send(packet)

	for {
		select {
		case p := <-packetCh:
			log.Printf("Received packet: %v\n", p)
		}
	}
}
