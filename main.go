package main

import (
  "flag"
  "fmt"
  "log"

  "github.com/kidoman/embd"
  _ "github.com/kidoman/embd/host/rpi"
)

const (
  // Read/write flags.
  WRITE_SINGLE_BYTE = 0x00
  WRITE_BURST = 0x40
  READ_SINGLE_BYTE = 0x80
  READ_BURST = 0xc0

  // Strobes
  SRES = 0x30
  SRX = 0x34

  // Status Registers
  PARTNUM = 0xf0
  VERSION = 0xf1

  // Config Registers
  IOCFG2 = 0x00
  IOCFG1 = 0x01
  IOCFG0 = 0x02

  FIFOTHR = 0x03

  SYNC1 = 0x04
  SYNC0 = 0x05

  PKTLEN = 0x06
  PKTCTRL1 = 0x07
  PKTCTRL0 = 0x08

  ADDR = 0x09

  CHANNR = 0x0a
  FSCTRL1 = 0x0b
  FSCTRL0 = 0x0c

  FREQ2 = 0x0d
  FREQ1 = 0x0e
  FREQ0 = 0x0f

  MDMCFG4 = 0x10
  MDMCFG3 = 0x11
  MDMCFG2 = 0x12
  MDMCFG1 = 0x13
  MDMCFG0 = 0x14

  DEVIATN = 0x15

  MCSM2 = 0x16
  MCSM1 = 0x17
  MCSM0 = 0x18

  FOCCFG = 0x19
  BSCFG = 0x1a

  AGCTRL2 = 0x1b
  AGCTRL1 = 0x1c
  AGCTRL0 = 0x1d

  WOREVT1 = 0x1e
  WOREVT0 = 0x1f
  WORCTRL = 0x20

  FREND1 = 0x21
  FREND0 = 0x22

  FSCAL3 = 0x23
  FSCAL2 = 0x24
  FSCAL1 = 0x25
  FSCAL0 = 0x26

  RCCTRL1 = 0x27
  RCCTRL0 = 0x28

  FSTEST = 0x29
  PTEST = 0x2a
  AGCTEST = 0x2b
  TEST2 = 0x2c
  TEST1 = 0x2d
  TEST0 = 0x2e
)

type CC1101 struct {
  bus embd.SPIBus
}

func (cc1101 *CC1101) Strobe(address byte) error {
  data := []byte{address, 0x00}
  return cc1101.bus.TransferAndReceiveData(data)
}

func (cc1101 *CC1101) ReadSingleByte(address byte) (byte, error) {
  data := []byte{address | READ_SINGLE_BYTE, 0x00}
  err := cc1101.bus.TransferAndReceiveData(data)
  if err != nil {
    return 0x00, err
  }
  return data[1], nil
}

func (cc1101 *CC1101) WriteSingleByte(address byte, in byte) error {
  data := []byte{address | WRITE_SINGLE_BYTE, in}
  return cc1101.bus.TransferAndReceiveData(data)
}

func (cc1101 *CC1101) Reset() error {
  return cc1101.Strobe(SRES)
}

func (c *CC1101) Init() error {
  c.WriteSingleByte(FSCTRL1, 0x08)
  c.WriteSingleByte(FSCTRL0, 0x00)

  c.SetCarrierFrequency(868)

  c.WriteSingleByte(MDMCFG4, 0x5b)
  c.WriteSingleByte(MDMCFG3, 0xf8)
  c.WriteSingleByte(MDMCFG2, 0x03)
  c.WriteSingleByte(MDMCFG1, 0x22)
  c.WriteSingleByte(MDMCFG0, 0xf8)

  c.WriteSingleByte(CHANNR, 0x00)
  c.WriteSingleByte(DEVIATN, 0x47)
  c.WriteSingleByte(FREND1, 0xb6)
  c.WriteSingleByte(FREND0, 0x10)
  c.WriteSingleByte(MCSM0, 0x18)
  c.WriteSingleByte(FOCCFG, 0x1d)
  c.WriteSingleByte(BSCFG, 0x1c)

  c.WriteSingleByte(AGCTRL2, 0xc7)
  c.WriteSingleByte(AGCTRL1, 0x00)
  c.WriteSingleByte(AGCTRL0, 0xb2)

  c.WriteSingleByte(FSCAL3, 0xea)
  c.WriteSingleByte(FSCAL2, 0x2a)
  c.WriteSingleByte(FSCAL1, 0x00)
  c.WriteSingleByte(FSCAL0, 0x11)

  c.WriteSingleByte(FSTEST, 0x59)
  c.WriteSingleByte(TEST2, 0x81)
  c.WriteSingleByte(TEST1, 0x35)
  c.WriteSingleByte(TEST0, 0x09)

  c.WriteSingleByte(IOCFG2, 0x0b)
  c.WriteSingleByte(IOCFG0, 0x06)

  // Two status bytes appended to payload: RSSI LQI and CRC OK.
  c.WriteSingleByte(PKTCTRL1, 0x04)
  // No address check, data whitening off, CRC enable, variable length packets.
  c.WriteSingleByte(PKTCTRL0, 0x05)

  c.WriteSingleByte(ADDR, 0x00)
  // Max packet length 61 bytes.
  c.WriteSingleByte(PKTLEN, 0x3d)
  return nil
}

func (cc1101 *CC1101) SelfTest() error {
  version, err := cc1101.ReadSingleByte(VERSION)
  if err != nil {
    return err
  }
  log.Printf("Version: 0x%x", version)
  partnum, err := cc1101.ReadSingleByte(PARTNUM)
  if err != nil {
    return err
  }
  log.Printf("Partnum: 0x%x", partnum)

  if version != 0x14 || partnum != 0x00 {
    return fmt.Errorf("Self test failed.\nGot Version: 0x%x Partnum: 0x%x", version, partnum)
  }
  return nil
}

func (cc1101 *CC1101) SetCarrierFrequency(freq int) error {
  if (freq == 868) {
    err := cc1101.WriteSingleByte(FREQ2, 0x21)
    if err != nil {
      return err
    }
    err = cc1101.WriteSingleByte(FREQ1, 0x62)
    if err != nil {
      return err
    }
    err = cc1101.WriteSingleByte(FREQ0, 0x76)
    if err != nil {
      return err
    }
    return nil
  } else {
    return fmt.Errorf("Frequency %dMHz not supported.", freq)
  }
}

func (cc1101 *CC1101) SetRx() error {
  return cc1101.Strobe(SRX)
}

func main() {
  flag.Parse()
  err := embd.InitSPI()
  if err != nil {
    panic(err)
  }
  defer embd.CloseSPI()

  bus := embd.NewSPIBus(embd.SPIMode0, 0, 50000, 8, 0)
  defer bus.Close()

  cc1101 := CC1101{
    bus: bus,
  }
  cc1101.Reset()
  cc1101.SelfTest()
  cc1101.Init()

  cc1101.SetRx()
}

