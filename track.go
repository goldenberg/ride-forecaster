package main

import (
	"fmt"
	"math"
	"sort"
	"time"

	stats "github.com/GaryBoone/GoStats/stats"
	geo "github.com/paulmach/go.geo"
	gpx "github.com/ptrv/go-gpx"
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

// PredictTrack converts a Path into a Track with the same number of points
// assuming a constant velocity and start time.
func PredictTrack(p *geo.Path, v Velocity, start time.Time) (t *Track) {
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

// Start returns the start time of the track.
func (t *Track) Start() time.Time {
	return t.times[0]
}

// End returns the end time of the track.
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

// Interpolate a Waypoint and associated Bearing at a given point in time along the track.
func (t *Track) Interpolate(mid time.Time) (*Waypoint, Bearing, float64, error) {
	if mid.Before(t.Start()) || mid.After(t.End()) {
		return nil, 0, 0, fmt.Errorf("time %s was before first time %s, or after last time %s", mid, t.End(), t.Start())
	}

	// Find the the closest points in time before and after the target time.
	endIdx := sort.Search(len(t.times), func(i int) bool { return t.times[i].After(mid) })
	startIdx := endIdx - 1

	end := t.times[endIdx]
	start := t.times[endIdx-1]

	// Range [0, 1] relative distance between two neighboring waypoints
	percent := float64(mid.Sub(start) / end.Sub(start))

	// And then interpolate the location at that point.
	line := geo.NewLine(t.path.GetAt(startIdx), t.path.GetAt(endIdx))
	midPt := line.Interpolate(percent)

	distance := t.GeoDistanceTo(startIdx) + t.path.GetAt(startIdx).GeoDistanceFrom(midPt)
	bearing := NewBearingFromDegrees(midPt.BearingTo(t.path.GetAt(endIdx)))
	// window := 25
	// bearing := t.LinearFitBearing(intMax(0, startIdx-window), intMin(endIdx+window, len(t.times)-1))

	return &Waypoint{midPt, mid}, bearing, distance, nil
}

func (t *Track) GeoDistanceTo(i int) (distance float64) {
	distance = 0.
	for j := 0; j < i; j++ {
		distance += t.path.GetAt(j).GeoDistanceFrom(t.path.GetAt(j + 1))
	}
	return distance
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

	// XXX: there's a dumb bug here, where this is always in the upper two quadrants. This makes the linear fit wrong!
	panic("currently, there's a bug where it's always in the upper two quadrants!")
	return NewBearing(math.Atan(r.Slope()))
}

func NewTrackFromGpxWpts(wpts []gpx.Wpt) (track *Track) {
	times := make([]time.Time, len(wpts), len(wpts))
	points := make([]geo.Point, len(wpts), len(wpts))

	for i, wpt := range wpts {
		points[i] = *geo.NewPoint(wpt.Lat, wpt.Lon)
		t := wpt.Time()
		// if t == 0 {
		// 	// log.Fatalf("Error '%s' parsing timestamp '%s'", err, wpt.Timestamp)
		// 	// XXXXXXX: really really stupid
		// 	times[i] = time.Now()
		// }
		times[i] = t
	}

	path := geo.NewPath()
	path.SetPoints(points)
	return &Track{path, times}
}
