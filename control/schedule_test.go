package control

import (
	"testing"

	"github.com/golang/protobuf/proto"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	scheduleA = &Schedule{
		Interval: []*Schedule_Interval{
			&Schedule_Interval{
				Begin: &Schedule_Time{
					Hour:   proto.Int32(9),
					Minute: proto.Int32(0),
				},
				End: &Schedule_Time{
					Hour:   proto.Int32(10),
					Minute: proto.Int32(30),
				},
				Type: Schedule_Interval_ON.Enum(),
			},
		},
	}

	scheduleB = &Schedule{
		Interval: []*Schedule_Interval{
			&Schedule_Interval{
				Begin: &Schedule_Time{
					Hour:   proto.Int32(9),
					Minute: proto.Int32(0),
				},
				End: &Schedule_Time{
					Hour:   proto.Int32(10),
					Minute: proto.Int32(30),
				},
				Type: Schedule_Interval_ON.Enum(),
			},
			&Schedule_Interval{
				Begin: &Schedule_Time{
					Hour:   proto.Int32(11),
					Minute: proto.Int32(0),
				},
				End: &Schedule_Time{
					Hour:   proto.Int32(12),
					Minute: proto.Int32(0),
				},
				Type: Schedule_Interval_OFF.Enum(),
			},
		},
	}
)

func TestBuildTree(t *testing.T) {
	Convey("Init", t, func() {
		tree := NewSchedule(scheduleA)
		So(tree, ShouldNotBeNil)
	})

	Convey("Contains", t, func() {
		tree := NewSchedule(scheduleB)
		So(tree, ShouldNotBeNil)
		So(tree.Query(9, 1), ShouldEqual, Schedule_Interval_ON)
		So(tree.Query(9, 0), ShouldEqual, Schedule_Interval_UNKNOWN)
		So(tree.Query(10, 29), ShouldEqual, Schedule_Interval_ON)
		So(tree.Query(10, 30), ShouldEqual, Schedule_Interval_UNKNOWN)
		So(tree.Query(11, 30), ShouldEqual, Schedule_Interval_OFF)
	})
}
