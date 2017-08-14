package sht31

import (
	"encoding/binary"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hatstand/shinywaffle/mocks"
	. "github.com/smartystreets/goconvey/convey"
)

func TestInit(t *testing.T) {
	mock := gomock.NewController(t)
	defer mock.Finish()
	i2c := mocks.NewMockI2CBus(mock)

	i2c.EXPECT().WriteByteToReg(byte(0x42), byte(0x30), byte(0xa2))

	sensor := NewSHT31(i2c, 0x42)
	sensor.Init()
}

func TestReadTemperatureAndHumidity(t *testing.T) {
	mock := gomock.NewController(t)
	defer mock.Finish()
	i2c := mocks.NewMockI2CBus(mock)

	rawTemp := make([]byte, 2)
	rawHum := make([]byte, 2)
	binary.BigEndian.PutUint16(rawTemp, (2<<15)-1)
	binary.BigEndian.PutUint16(rawHum, (2<<15)-1)
	gomock.InOrder(
		i2c.EXPECT().WriteByteToReg(byte(0x42), byte(0x24), byte(0x00)),
		i2c.EXPECT().ReadBytes(byte(0x42), 6).Return(
			[]byte{
				rawTemp[0], rawTemp[1], crc(rawTemp),
				rawHum[0], rawHum[1], crc(rawHum),
			}, nil),
	)

	sensor := NewSHT31(i2c, 0x42)
	Convey("Reads temperature", t, func() {
		temp, humidity, err := sensor.ReadTempAndHum()
		So(err, ShouldBeNil)
		So(temp, ShouldEqual, 130)
		So(humidity, ShouldEqual, 100)
	})
}
