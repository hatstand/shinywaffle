package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"
)

const (
	maxDatagramSize = 8192
)

var discover = []byte("{\"system\":{\"get_sysinfo\":null},\"emeter\":{\"get_realtime\":null}}")

type SmartPlugMessage struct {
	Emeter Emeter `json:"emeter"`
	System System `json:"system"`
}

type Emeter struct {
	Realtime Realtime `json:"get_realtime"`
}

type System struct {
	SysInfo SysInfo `json:"get_sysinfo"`
}

type Realtime struct {
	Current float32
	Error   int `json:"err_code"`
	Power   float32
	Total   float32
	Voltage float32
}

type SysInfo struct {
	Mode       string `json:"active_mode"`
	Name       string `json:"alias"`
	DeviceName string `json:"dev_name"`
	ID         string `json:"deviceId"`
	Error      int    `json:"err_code"`
	MAC        string
	Model      string
	State      int `json:"relay_state"`
}

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
		fmt.Println(string(deobfuscate(packet)))
		var message SmartPlugMessage
		err = json.Unmarshal(deobfuscate(packet), &message)
		if err != nil {
			log.Fatal("Failed to decode as JSON")
		}
		out, _ := json.MarshalIndent(message, "", "  ")
		fmt.Printf("%s\n", out)
	}
}
