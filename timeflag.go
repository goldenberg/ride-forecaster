package main

import (
	"time"
)

// -- time.Time Value
type timeValue time.Time

func newTimeValue(val time.Time, p *time.Time) *timeValue {
	*p = val
	return (*timeValue)(p)
}

const DEFAULT_STAMP_FORMAT = "01/02/06 15:04"

func (t *timeValue) Set(s string) error {
	// shouldn't hard code this. should get from start location of GPX
	LosAngeles, _ := time.LoadLocation("America/Los_Angeles")
	v, err := time.ParseInLocation(LosAngeles, s, PST)
	if err != nil {
		return err
	}
	*t = timeValue(v)
	return nil
}

func (t *timeValue) Get() time.Time { return time.Time(*t) }

func (t *timeValue) String() string {
	return t.Get().Format(time.Stamp)
}
