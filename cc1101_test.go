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
