package shinywaffle

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hatstand/shinywaffle/mocks"

	. "github.com/smartystreets/goconvey/convey"
)

func WithMocks(t *testing.T, f func(bus *mocks.MockSPIBus, cc1101 *CC1101)) func() {
	return func() {
		mock := gomock.NewController(t)
		defer mock.Finish()
		bus := mocks.NewMockSPIBus(mock)
		cc1101 := &CC1101{bus: bus}
		f(bus, cc1101)
	}
}

func TestSelfTest(t *testing.T) {
	Convey("Init", t, WithMocks(t, func(bus *mocks.MockSPIBus, cc1101 *CC1101) {
		bus.EXPECT().TransferAndReceiveData([]byte{VERSION | READ_SINGLE_BYTE, 0x00}).Return(nil).SetArg(0, []byte{0x00, 0x14})
		bus.EXPECT().TransferAndReceiveData([]byte{PARTNUM | READ_SINGLE_BYTE, 0x00}).Return(nil).SetArg(0, []byte{0x00, 0x00})

		So(cc1101.SelfTest(), ShouldBeNil)
	}))
}

func TestStrobe(t *testing.T) {
	Convey("Strobe", t, WithMocks(t, func(bus *mocks.MockSPIBus, cc1101 *CC1101) {
		bus.EXPECT().TransferAndReceiveData([]byte{0x42, 0x00}).Return(nil).SetArg(0, []byte{0x43, 0x00})

		ret, err := cc1101.Strobe(0x42)
		So(err, ShouldBeNil)
		So(ret, ShouldEqual, 0x43)
	}))
}

func TestReset(t *testing.T) {
	Convey("Reset", t, WithMocks(t, func(bus *mocks.MockSPIBus, cc1101 *CC1101) {
		bus.EXPECT().TransferAndReceiveData([]byte{SRES, 0x00}).Return(nil)

		So(cc1101.Reset(), ShouldBeNil)
	}))
}

func TestSetState(t *testing.T) {
	Convey("RX", t, WithMocks(t, func(bus *mocks.MockSPIBus, cc1101 *CC1101) {
		bus.EXPECT().TransferAndReceiveData([]byte{SRX, 0x00}).Return(nil)
		So(cc1101.SetRx(), ShouldBeNil)
	}))
	Convey("TX", t, WithMocks(t, func(bus *mocks.MockSPIBus, cc1101 *CC1101) {
		bus.EXPECT().TransferAndReceiveData([]byte{STX, 0x00}).Return(nil)
		So(cc1101.SetTx(), ShouldBeNil)
	}))
	Convey("IDLE", t, WithMocks(t, func(bus *mocks.MockSPIBus, cc1101 *CC1101) {
		bus.EXPECT().TransferAndReceiveData([]byte{SIDLE, 0x00}).Return(nil)
		So(cc1101.SetIdle(), ShouldBeNil)
	}))
	Convey("Flush RX buffer", t, WithMocks(t, func(bus *mocks.MockSPIBus, cc1101 *CC1101) {
		gomock.InOrder(
			bus.EXPECT().TransferAndReceiveData([]byte{SIDLE, 0x00}).Return(nil),
			bus.EXPECT().TransferAndReceiveData([]byte{SFRX, 0x00}).Return(nil),
		)
		cc1101.FlushRx()
	}))
}
