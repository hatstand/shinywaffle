package control

import (
	"github.com/hatstand/shinywaffle"
)

const (
	Day     = 0x05
	Night   = 0x03
	Defrost = 0x09
	Off     = 0x60
	Auto    = 0x11
)

func convertTemp(temp float32) byte {
	return byte(int(temp*10) * 2 / 10)
}

type Controller struct {
	radio shinywaffle.CC1101
}

func NewController() *Controller {
	return &Controller{
		radio: shinywaffle.NewCC1101(nil),
	}
}

func (c *Controller) TurnOn(addr []byte) {
	packet := []byte{0x57, 0x16, 0x0a}
	packet = append(packet, addr[0])
	packet = append(packet, addr[1])
	packet = append(packet, Day)
	packet = append(packet, convertTemp(30))
	packet = append(packet, convertTemp(30))
	packet = append(packet, convertTemp(10))
	c.radio.Send(packet)
	c.radio.Send(packet)
	c.radio.Send(packet)
}

func (c *Controller) TurnOff(addr []byte) {
	packet := []byte{0x57, 0x16, 0x0a}
	packet = append(packet, addr[0])
	packet = append(packet, addr[1])
	packet = append(packet, Defrost)
	packet = append(packet, convertTemp(30))
	packet = append(packet, convertTemp(30))
	packet = append(packet, convertTemp(10))
	c.radio.Send(packet)
	c.radio.Send(packet)
	c.radio.Send(packet)
}
