package shinywaffle

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hatstand/shinywaffle/mocks"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSelfTest(t *testing.T) {
	Convey("Init", t, func() {
		mock := gomock.NewController(t)
		defer mock.Finish()
		bus := mocks.NewMockSPIBus(mock)
		cc1101 := CC1101{bus: bus}

		bus.EXPECT().TransferAndReceiveData([]byte{VERSION | READ_SINGLE_BYTE, 0x00}).Return(nil).SetArg(0, []byte{0x00, 0x14})
		bus.EXPECT().TransferAndReceiveData([]byte{PARTNUM | READ_SINGLE_BYTE, 0x00}).Return(nil).SetArg(0, []byte{0x00, 0x00})

		So(cc1101.SelfTest(), ShouldBeNil)
	})
}
