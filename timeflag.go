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
	// PST, err := time.LoadLocation("PST")
	v, err := time.Parse(DEFAULT_STAMP_FORMAT, s)
	if err != nil {
		return err
	}
	*t = timeValue(v)
	return nil
}

func (t *timeValue) Get() time.Time { return time.Time(*t) }

func (t *timeValue) String() string {
	return "foobar"
	// return &t.Format(time.Stamp)
}
