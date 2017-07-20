package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/hatstand/shinywaffle"
)

const (
	SyncWord = 0x4242
)

func main() {
	packetCh := make(chan []byte, 10)
	cc1101 := shinywaffle.NewCC1101(packetCh)
	defer cc1101.Close()
	cc1101.SetSyncWord(SyncWord)

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)

	for {
		select {
		case p := <-packetCh:
			log.Printf("Received packet: %v\n", p)
		case <-signalCh:
			return
		}
	}
}
