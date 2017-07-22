package sht31

import (
	"encoding/binary"
	"fmt"
	"sync"
	"time"

	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/rpi"
	"github.com/sigurn/crc8"
)

const (
	DEFAULT_ADDR = 0x44
	READ_STATUS  = 0xf32d
	CLEAR_STATUS = 0x3041
	SOFT_RESET   = 0x30a2

	MEAS_HIGHREP = 0x2400
)

var (
	// See SHT3x datasheet; section 4.12.
	CRC8_PARAMS = crc8.Params{0x31, 0xff, false, false, 0x00, 0x92, "foo"}
	CRC8_TABLE  = crc8.MakeTable(CRC8_PARAMS)
)

func crc(data []byte) byte {
	return crc8.Checksum(data, CRC8_TABLE)
}

func convertTemperature(raw uint16) float32 {
	return -45 + 175*(float32(raw)/((2<<15)-1))
}

func convertHumidity(raw uint16) float32 {
	return 100 * (float32(raw) / ((2 << 15) - 1))
}

type SHT31 struct {
	bus     embd.I2CBus
	address byte
	lock    sync.Mutex
}

func NewSHT31(bus embd.I2CBus, address byte) *SHT31 {
	return &SHT31{
		bus:     bus,
		address: address,
	}
}

func (s *SHT31) Init() {
	s.writeCommand(SOFT_RESET)
	time.Sleep(100 * time.Millisecond)
}

func (s *SHT31) writeCommand(command uint16) error {
	return s.bus.WriteByteToReg(s.address, byte(command>>8), byte(command&0xff))
}

func (s *SHT31) ReadStatus() (uint16, error) {
	s.writeCommand(READ_STATUS)
	status, _ := s.bus.ReadBytes(s.address, 3)
	if crc(status[:2]) != status[2] {
		return 0, fmt.Errorf("CRC check failed for status")
	}
	return binary.BigEndian.Uint16(status[:2]), nil
}

func (s *SHT31) ReadTempAndHum() (float32, float32, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	err := s.writeCommand(MEAS_HIGHREP)
	if err != nil {
		return 0, 0, fmt.Errorf("Failed to read temperature: %v", err)
	}
	time.Sleep(500 * time.Millisecond)
	value, err := s.bus.ReadBytes(s.address, 6)
	if err != nil {
		return 0, 0, fmt.Errorf("Failed to read temperature: %v", err)
	}

	if value[2] != crc(value[:2]) {
		return 0, 0, fmt.Errorf("CRC check failed for temperature")
	}

	if value[5] != crc(value[3:5]) {
		return 0, 0, fmt.Errorf("CRC check failed for humidity")
	}

	rawTemp := binary.BigEndian.Uint16(value[:2])
	rawHumidity := binary.BigEndian.Uint16(value[3:5])
	return convertTemperature(rawTemp), convertHumidity(rawHumidity), nil
}
