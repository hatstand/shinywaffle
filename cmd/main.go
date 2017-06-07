package main

import (
  "flag"
  "log"
  "time"

  "github.com/hatstand/cc1101"
)

func main() {
  flag.Parse()

  packetCh := make(chan []byte, 10)
  cc1101 := cc1101.NewCC1101(packetCh)
  defer cc1101.Close()

  time.Sleep(5 * time.Second)
  cc1101.Send([]byte{0x57, 0x16, 0x0a, 0x2b, 0x7e, 0x60, 0x3b, 0x26, 0x14})
  cc1101.Send([]byte{0x57, 0x16, 0x0a, 0x2b, 0x7e, 0x60, 0x3b, 0x26, 0x14})
  cc1101.Send([]byte{0x57, 0x16, 0x0a, 0x2b, 0x7e, 0x60, 0x3b, 0x26, 0x14})

  for {
    select{
      case p := <-packetCh:
        log.Printf("Received packet: %v\n", p)
    }
  }
}

