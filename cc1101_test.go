package shinywaffle

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hatstand/shinywaffle/mocks"

	. "github.com/smartystreets/goconvey/convey"
)

func WithBusAndPins(t *testing.T, f func(bus *mocks.MockSPIBus, gdo0 *mocks.MockDigitalPin, gdo2 *mocks.MockDigitalPin, cc1101 *CC1101)) func() {
	return func() {
		mock := gomock.NewController(t)
		defer mock.Finish()
		bus := mocks.NewMockSPIBus(mock)
		gdo0 := mocks.NewMockDigitalPin(mock)
		gdo2 := mocks.NewMockDigitalPin(mock)
		cc1101 := &CC1101{
			bus:  bus,
			gdo0: gdo0,
			gdo2: gdo2,
		}
		f(bus, gdo0, gdo2, cc1101)
	}
}

func WithBus(t *testing.T, f func(bus *mocks.MockSPIBus, cc1101 *CC1101)) func() {
	return WithBusAndPins(t, func(bus *mocks.MockSPIBus, gdo0 *mocks.MockDigitalPin, gdo2 *mocks.MockDigitalPin, cc1101 *CC1101) {
		f(bus, cc1101)
	})
}

func TestSelfTest(t *testing.T) {
	Convey("Init", t, WithBus(t, func(bus *mocks.MockSPIBus, cc1101 *CC1101) {
		expectTransferAndReceive(bus, []byte{VERSION | READ_SINGLE_BYTE, 0x00}, []byte{0x00, 0x14})
		expectTransferAndReceive(bus, []byte{PARTNUM | READ_SINGLE_BYTE, 0x00}, []byte{0x00, 0x00})

		So(cc1101.SelfTest(), ShouldBeNil)
	}))
}

func TestStrobe(t *testing.T) {
	Convey("Strobe", t, WithBus(t, func(bus *mocks.MockSPIBus, cc1101 *CC1101) {
		expectTransferAndReceive(bus, []byte{0x42, 0x00}, []byte{0x43, 0x00})

		ret, err := cc1101.Strobe(0x42)
		So(err, ShouldBeNil)
		So(ret, ShouldEqual, 0x43)
	}))
}

func TestReset(t *testing.T) {
	Convey("Reset", t, WithBus(t, func(bus *mocks.MockSPIBus, cc1101 *CC1101) {
		bus.EXPECT().TransferAndReceiveData([]byte{SRES, 0x00}).Return(nil)

		So(cc1101.Reset(), ShouldBeNil)
	}))
}

func TestSetState(t *testing.T) {
	Convey("RX", t, WithBus(t, func(bus *mocks.MockSPIBus, cc1101 *CC1101) {
		bus.EXPECT().TransferAndReceiveData([]byte{SRX, 0x00}).Return(nil)
		So(cc1101.SetRx(), ShouldBeNil)
	}))
	Convey("TX", t, WithBus(t, func(bus *mocks.MockSPIBus, cc1101 *CC1101) {
		bus.EXPECT().TransferAndReceiveData([]byte{STX, 0x00}).Return(nil)
		So(cc1101.SetTx(), ShouldBeNil)
	}))
	Convey("IDLE", t, WithBus(t, func(bus *mocks.MockSPIBus, cc1101 *CC1101) {
		bus.EXPECT().TransferAndReceiveData([]byte{SIDLE, 0x00}).Return(nil)
		So(cc1101.SetIdle(), ShouldBeNil)
	}))
	Convey("Flush RX buffer", t, WithBus(t, func(bus *mocks.MockSPIBus, cc1101 *CC1101) {
		gomock.InOrder(
			bus.EXPECT().TransferAndReceiveData([]byte{SIDLE, 0x00}).Return(nil),
			bus.EXPECT().TransferAndReceiveData([]byte{SFRX, 0x00}).Return(nil),
		)
		cc1101.FlushRx()
	}))
}

func TestSendPacket(t *testing.T) {
	Convey("Send", t, WithBusAndPins(t, func(bus *mocks.MockSPIBus, gdo0 *mocks.MockDigitalPin, gdo2 *mocks.MockDigitalPin, cc1101 *CC1101) {
		packet := []byte{0x42, 0x43, 0x44}
		gomock.InOrder(
			// Write the packet length.
			bus.EXPECT().TransferAndReceiveData([]byte{TXFIFO | WRITE_SINGLE_BYTE, 3}),
			// Write the packet data.
			bus.EXPECT().TransferAndReceiveData([]byte{TXFIFO | WRITE_BURST, 0x42, 0x43, 0x44}),
			// Switch to send mode.
			bus.EXPECT().TransferAndReceiveData([]byte{STX, 0x00}),
			// Wait for sync word to transmit (rising edge on gdo2).
			gdo2.EXPECT().Read().Return(1, nil),
			// Wait for packet data to transmit (falling edge on gdo2).
			gdo2.EXPECT().Read().Return(0, nil),
			// Switch back to idle mode when finished.
			bus.EXPECT().TransferAndReceiveData([]byte{SIDLE, 0x00}),
			// Flush the TX buffer.
			bus.EXPECT().TransferAndReceiveData([]byte{SFTX, 0x00}),
			// Return to RX mode.
			bus.EXPECT().TransferAndReceiveData([]byte{SRX, 0x00}),
		)
		cc1101.Send(packet)
	}))
}

func TestReceivePacket(t *testing.T) {
	Convey("Send", t, WithBusAndPins(t, func(bus *mocks.MockSPIBus, gdo0 *mocks.MockDigitalPin, gdo2 *mocks.MockDigitalPin, cc1101 *CC1101) {
		addr := byte(RXFIFO | READ_BURST)
		packet := []byte{0x42, 0x43, 0x44}
		response := []byte{0x00}
		response = append(response, packet...)
		gomock.InOrder(
			// Read RXBYTES for fifo length and overflow.
			expectTransferAndReceive(bus, []byte{RXBYTES | READ_SINGLE_BYTE, 0x00}, []byte{0x00, 0x03}),
			// Read first RXFIFO byte for packet length.
			expectTransferAndReceive(bus, []byte{RXFIFO | READ_SINGLE_BYTE, 0x00}, []byte{0x00, 0x03}),
			// Read packet data out of RXFIFO.
			expectTransferAndReceive(bus, []byte{
				addr,
				addr + 1*8 | READ_BURST,
				addr + 2*8 | READ_BURST,
				addr + 3*8 | READ_BURST,
			}, response),
			// Read packet status bytes.
			bus.EXPECT().TransferAndReceiveData([]byte{
				addr,
				addr + 1*8 | READ_BURST,
				addr + 2*8 | READ_BURST,
			}),
			// Flush RX buffer.
			bus.EXPECT().TransferAndReceiveData([]byte{SIDLE, 0x00}),
			bus.EXPECT().TransferAndReceiveData([]byte{SFRX, 0x00}),
		)
		recv, err := cc1101.Receive()
		So(err, ShouldBeNil)
		So(recv, ShouldResemble, packet)
	}))
}

func expectTransferAndReceive(bus *mocks.MockSPIBus, in []byte, out []byte) *gomock.Call {
	return bus.EXPECT().TransferAndReceiveData(in).Return(nil).Do(func(in []byte) {
		copy(in, out)
	})
}
