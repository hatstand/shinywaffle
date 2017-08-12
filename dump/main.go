package main

import (
	"encoding/hex"
	"fmt"

	"github.com/hatstand/shinywaffle"
)

func dumpPacket(packet []byte) {
	fmt.Printf("%s\n", hex.EncodeToString(packet))
}

func main() {
	packetCh := make(chan []byte, 10)
	cc1101 := shinywaffle.NewCC1101(packetCh)
	defer cc1101.Close()
	cc1101.SetSyncWord(0xd391)

	for {
		select {
		case p := <-packetCh:
			dumpPacket(p)
		}
	}
}
