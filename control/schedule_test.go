package control

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	scheduleA = &Schedule{
		Interval: []*Schedule_Interval{
			&Schedule_Interval{
				Begin: &Schedule_Time{
					Hour:   9,
					Minute: 0,
				},
				End: &Schedule_Time{
					Hour:   10,
					Minute: 30,
				},
				TargetTemperature: 15,
			},
		},
	}

	scheduleB = &Schedule{
		Interval: []*Schedule_Interval{
			&Schedule_Interval{
				Begin: &Schedule_Time{
					Hour:   9,
					Minute: 0,
				},
				End: &Schedule_Time{
					Hour:   10,
					Minute: 30,
				},
				TargetTemperature: 12,
			},
			&Schedule_Interval{
				Begin: &Schedule_Time{
					Hour:   11,
					Minute: 0,
				},
				End: &Schedule_Time{
					Hour:   12,
					Minute: 0,
				},
				TargetTemperature: 42,
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
		So(tree.Query(9, 1), ShouldEqual, 12)
		So(tree.Query(9, 0), ShouldEqual, -1)
		So(tree.Query(10, 29), ShouldEqual, 12)
		So(tree.Query(10, 30), ShouldEqual, -1)
		So(tree.Query(11, 30), ShouldEqual, 42)
	})
}
