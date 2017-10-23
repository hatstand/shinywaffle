package control

import (
	"time"
	"unsafe"

	"github.com/golang-collections/go-datastructures/augmentedtree"
)

type ScheduleInterval struct {
	Start int64
	End   int64
	State Schedule_Interval_State
}

func (s *ScheduleInterval) LowAtDimension(uint64) int64 {
	return s.Start
}

func (s *ScheduleInterval) HighAtDimension(uint64) int64 {
	return s.End
}

func (s *ScheduleInterval) OverlapsAtDimension(other augmentedtree.Interval, d uint64) bool {
	return other.LowAtDimension(d) <= s.HighAtDimension(d) && other.HighAtDimension(d) >= s.LowAtDimension(d)
}

func (s *ScheduleInterval) ID() uint64 {
	return *(*uint64)(unsafe.Pointer(s))
}

type queryInterval struct {
	start int64
	end   int64
}

func (q *queryInterval) LowAtDimension(uint64) int64 {
	return q.start
}

func (q *queryInterval) HighAtDimension(uint64) int64 {
	return q.end
}

func (q *queryInterval) OverlapsAtDimension(other augmentedtree.Interval, d uint64) bool {
	return other.LowAtDimension(d) <= q.HighAtDimension(d) && other.HighAtDimension(d) >= q.LowAtDimension(d)
}

func (q *queryInterval) ID() uint64 {
	return 0
}

func newInterval(proto *Schedule_Interval) augmentedtree.Interval {
	return &ScheduleInterval{
		Start: int64(proto.GetBegin().GetHour()*60 + proto.GetBegin().GetMinute()),
		End:   int64(proto.GetEnd().GetHour()*60 + proto.GetEnd().GetMinute()),
		State: proto.GetType(),
	}
}

type IntervalTree struct {
	tree augmentedtree.Tree
}

func (t *IntervalTree) Query(hour int, minute int) Schedule_Interval_State {
	is := t.tree.Query(&queryInterval{
		start: int64(hour*60 + minute),
		end:   int64(hour*60 + minute),
	})
	if len(is) == 0 {
		return Schedule_Interval_UNKNOWN
	}
	return is[0].(*ScheduleInterval).State
}

func (t *IntervalTree) QueryTime(ti time.Time) Schedule_Interval_State {
	return t.Query(ti.Hour(), ti.Minute())
}

func (t *IntervalTree) FetchDay() []*ScheduleInterval {
	is := t.tree.Query(&queryInterval{
		start: 0,
		end:   60 * 24,
	})
	ret := make([]*ScheduleInterval, len(is))
	for i, _ := range is {
		ret[i] = is[i].(*ScheduleInterval)
	}
	return ret
}

func NewSchedule(proto *Schedule) *IntervalTree {
	tree := augmentedtree.New(1)
	if proto != nil {
		for _, i := range proto.Interval {
			tree.Add(newInterval(i))
		}
	}
	return &IntervalTree{
		tree: tree,
	}
}
