package main

import (
	"fmt"
	stats "github.com/GaryBoone/GoStats/stats"
	geo "github.com/paulmach/go.geo"
	gpx "github.com/ptrv/go-gpx"
	"math"
	"sort"
	"time"
)

type Track struct {
	path  *geo.Path
	times []time.Time
}

// NewTrack creates a new track from a path and slice of times.
func NewTrack(p *geo.Path, times []time.Time) (*Track, error) {
	if p.Length() != len(times) {
		return nil, fmt.Errorf("path had length %i but times had length %i. Must be equal.",
			p.Length(),
			len(times))
	}
	return &Track{p, times}, nil
}

// PredictTrack converts a Path into a Track assuming a constant velocity and start time.
func PredictTrack(p *geo.Path, v Velocity, start time.Time) (t *Track) {
	// var pathDist = p.GeoDistance()
	var n = p.Length()
	var times = make([]time.Time, n, n)

	times[0] = start

	for i := 0; i < n-1; i++ {
		var segmentDist = p.GetAt(i).GeoDistanceFrom(p.GetAt(i+1), true)
		// XXX: This can't possibly be idiomatic but it seems to work
		var timeDelta = time.Duration(segmentDist / v.Ms() * float64(time.Second))
		times[i+1] = times[i].Add(timeDelta)
	}
	t, err := NewTrack(p, times)
	if err != nil {
		panic("times was created to be the same length. Should be impossible")
	}
	return t
}

func (t *Track) Path() *geo.Path {
	return t.path
}

func (t *Track) Start() time.Time {
	return t.times[0]
}

func (t *Track) End() time.Time {
	return t.times[len(t.times)-1]
}

// TimeShift translates the Track to a different start time.
func (t *Track) TimeShift(newStart time.Time) *Track {
	var newTimes = make([]time.Time, len(t.times), len(t.times))
	var delta = newStart.Sub(t.times[0])

	for i := 0; i < len(t.times); i++ {
		newTimes[i] = t.times[i].Add(delta)
	}
	shifted, err := NewTrack(t.path, newTimes)
	if err != nil {
		panic("times was created to be the same length. Should be impossible")
	}
	return shifted
}

// Waypoint returns the waypoint at a given index
func (t *Track) Waypoint(i int) *Waypoint {
	return &Waypoint{t.path.GetAt(i), t.times[i]}
}

func (t *Track) WayPointAndBearingAtTime(mid time.Time) (*Waypoint, Bearing, error) {
	if mid.Before(t.times[0]) || mid.After(t.times[len(t.times)-1]) {
		return nil, 0, fmt.Errorf("time %s was before first time %s, or after last time %s", mid, t.times[0], t.times[len(t.times)-1])
	}

	endIdx := sort.Search(len(t.times), func(i int) bool { return t.times[i].After(mid) })
	startIdx := endIdx - 1

	end := t.times[endIdx]
	start := t.times[endIdx-1]

	// Range [0, 1] relative distance between two neighboring waypoints
	percent := float64(mid.Sub(start) / end.Sub(start))
	line := geo.NewLine(t.path.GetAt(startIdx), t.path.GetAt(endIdx))
	midPt := line.Interpolate(percent)

	bearing := NewBearingFromDegrees(midPt.BearingTo(t.path.GetAt(endIdx)))
	// window := 25
	// bearing := t.LinearFitBearing(intMax(0, startIdx-window), intMin(endIdx+window, len(t.times)-1))
	return &Waypoint{midPt, mid}, bearing, nil
}

func intMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func intMax(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (t *Track) LinearFitBearing(start, end int) Bearing {
	var r stats.Regression
	for i := start; i <= end; i++ {
		pt := t.Waypoint(i)
		r.Update(pt.X(), pt.Y())
	}

	// XXX: there's a dumb bug here, where this is always in the upper two quadrants
	return NewBearing(math.Atan(r.Slope()))
}

func NewTrackFromGpxWpts(wpts []gpx.GpxWpt) (track *Track) {
	times := make([]time.Time, len(wpts), len(wpts))
	points := make([]geo.Point, len(wpts), len(wpts))

	for i, wpt := range wpts {
		points[i] = *geo.NewPoint(wpt.Lat, wpt.Lon)
		t, err := time.Parse(gpx.TIMELAYOUT, wpt.Timestamp)
		if err != nil {
			// log.Fatalf("Error '%s' parsing timestamp '%s'", err, wpt.Timestamp)
			// XXXXXXX: really really stupid
			times[i] = time.Now()
		}
		times[i] = t
	}

	path := geo.NewPath()
	path.SetPoints(points)
	return &Track{path, times}
}
