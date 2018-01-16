package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"time"
)

const (
	maxDatagramSize = 8192
)

var discover = []byte("{\"system\":{\"get_sysinfo\":null},\"emeter\":{\"get_realtime\":null}}")

func obfuscate(data []byte) []byte {
	k := byte(171)
	ret := make([]byte, len(data))
	for i := 0; i < len(data); i++ {
		ret[i] = data[i] ^ k
		k = ret[i]
	}
	return ret
}

func deobfuscate(data []byte) []byte {
	k := byte(171)
	ret := make([]byte, len(data))
	for i := 0; i < len(data); i++ {
		ret[i] = data[i] ^ k
		k = data[i]
	}
	return ret
}

func main() {
	addr, err := net.ResolveUDPAddr("udp", "255.255.255.255:9999")
	if err != nil {
		log.Fatal("Failed to resolve UDP addr:", err)
	}

	conn, err := net.ListenUDP("udp", nil)
	packetCh := make(chan []byte)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	go func() {
		for {
			buf := make([]byte, maxDatagramSize)
			conn.SetReadDeadline(time.Now().Add(10 * time.Millisecond))
			n, _, _ := conn.ReadFromUDP(buf)
			if n > 0 {
				packetCh <- buf[:n]
			}
			if ctx.Err() != nil {
				close(packetCh)
				return
			}
		}
	}()
	conn.WriteTo(obfuscate(discover), addr)
	for packet := range packetCh {
		fmt.Println(hex.Dump(deobfuscate(packet)))
	}
}
