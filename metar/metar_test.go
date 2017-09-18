package metar

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestParseTemperatures(t *testing.T) {
	Convey("Temperature", t, func() {
		So(parseTemperature("42"), ShouldEqual, 42)
		So(parseTemperature("M03"), ShouldEqual, -3)
	})
}

func TestParseReportType(t *testing.T) {
	Convey("ReportType", t, func() {
		So(parseReportType("METAR"), ShouldEqual, Routine)
		So(parseReportType("TAF AMD"), ShouldEqual, Forecast)
	})
}

func TestParseMETARs(t *testing.T) {
	Convey("ParseMETARs", t, func() {
		metars, err := ParseMETARs("201708312350 METAR EGLC 312350Z AUTO 24004KT 9999 NCD 14/10 Q1020=\n")
		So(err, ShouldBeNil)
		So(metars, ShouldNotBeEmpty)
		m := metars[0]
		So(m.ReportType, ShouldEqual, Routine)
		So(m.ICAO, ShouldEqual, "EGLC")
		So(m.Temperature, ShouldEqual, 14)
		So(m.DewPoint, ShouldEqual, 10)
		So(m.DateTime.Year(), ShouldEqual, 2017)
		So(m.DateTime.Month(), ShouldEqual, 8)
		So(m.DateTime.Day(), ShouldEqual, 31)
		So(m.DateTime.Hour(), ShouldEqual, 23)
		So(m.DateTime.Minute(), ShouldEqual, 50)
	})

	Convey("ParseEGSS", t, func() {
		metars, err := ParseMETARs("201709081620 METAR COR EGSS 081620Z 24009KT 9000 SHRA BKN049CB 15/13 Q0996=\n")
		So(err, ShouldBeNil)
		So(metars, ShouldNotBeEmpty)
		m := metars[0]
		So(m.ReportType, ShouldEqual, Routine)
		So(m.ICAO, ShouldEqual, "EGSS")
		So(m.Temperature, ShouldEqual, 15)
		So(m.DewPoint, ShouldEqual, 13)
		So(m.DateTime.Year(), ShouldEqual, 2017)
		So(m.DateTime.Month(), ShouldEqual, 9)
		So(m.DateTime.Day(), ShouldEqual, 8)
		So(m.DateTime.Hour(), ShouldEqual, 16)
		So(m.DateTime.Minute(), ShouldEqual, 20)
	})
}
