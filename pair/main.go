package main

import (
	"encoding/hex"
	"flag"
	"log"
	"time"

	"github.com/hatstand/shinywaffle"
)

var address = flag.String("address", "", "Address in hexadecimal to pair as")

func main() {
	flag.Parse()

	addr, err := hex.DecodeString(*address)
	if err != nil || len(addr) != 2 {
		log.Fatal("Address must be exactly 2 bytes in hexadecimal")
	}

	packetCh := make(chan []byte, 10)
	cc1101 := shinywaffle.NewCC1101(packetCh)
	defer cc1101.Close()
	defer close(packetCh)
	cc1101.SetSyncWord(0xd391)

	// Pairing packet
	packet := []byte{0x57, 0x96, 0x0a}
	// Set the address we want the radiator to have.
	packet = append(packet, addr[0])
	packet = append(packet, addr[1])
	// Set some sensible defaults for initial radiator settings.
	packet = append(packet, 0x60) // Off
	packet = append(packet, 40)   // 20C Day time
	packet = append(packet, 30)   // 15C Night time
	packet = append(packet, 20)   // 10C Defrost

	log.Printf("Pairing as: %s\n", *address)
	for i := 0; i < 10; i++ {
		cc1101.Send(packet)
		cc1101.Send(packet)
		cc1101.Send(packet)
		time.Sleep(1 * time.Second)
	}
}
