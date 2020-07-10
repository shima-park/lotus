package pipeliner

import (
	"time"
)

var (
	defaultScheduleParser ScheduleParser = func(string) (Schedule, error) {
		return defaultSchedule, nil
	}
	defaultSchedule Schedule = ConstantDelaySchedule{}
)

type ScheduleParser func(spec string) (Schedule, error)

type Schedule interface {
	// Next returns the next activation time, later than the given time.
	// Next is invoked initially, and then each time the job is run.
	Next(time.Time) time.Time
}

type ConstantDelaySchedule struct {
}

func (schedule ConstantDelaySchedule) Next(t time.Time) time.Time {
	return t
}
