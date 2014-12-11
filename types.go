package main

import (
	"fmt"
	geo "github.com/paulmach/go.geo"
	gpx "github.com/ptrv/go-gpx"
	// "log"
	"math"
	"time"
)

type Waypoint struct {
	*geo.Point
	Time time.Time
}

type Track struct {
	path  *geo.Path
	times []time.Time
}

func (t *Track) Path() *geo.Path {
	return t.path
}

// Velocity is measured in m/s
type Velocity float64

func NewVelocityFromMph(mph float64) Velocity {
	return Velocity(mph * 1609.34 / 3600.)

}
func (v Velocity) Mph() float64 {
	return float64(v/1609.34) * 3600
}

func (v Velocity) Ms() float64 {
	return float64(v)
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

// NewTrack creates a new track from a path and slice of
func NewTrack(p *geo.Path, times []time.Time) (*Track, error) {
	if p.Length() != len(times) {
		return nil, fmt.Errorf("path had length %i but times had length %i. Must be equal.",
			p.Length(),
			len(times))
	}
	return &Track{p, times}, nil
}

// Waypoint returns the waypoint at a given index
func (t *Track) Waypoint(i int) *Waypoint {
	return &Waypoint{t.path.GetAt(i), t.times[i]}
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

func deg2Rad(d float64) float64 {
	return d * math.Pi / 180.0
}

func rad2Deg(r float64) float64 {
	return 180.0 * r / math.Pi
}

func makeRadPos(r float64) float64 {
	if r < 0 {
		return 2*math.Pi - r
	}
	return r
}

type Bearing float64

func NewBearing(r float64) Bearing {
	return Bearing(r).Normalize()
}

func NewBearingFromDegrees(d float64) Bearing {
	return Bearing(deg2Rad(d)).Normalize()
}

func (b Bearing) Normalize() Bearing {
	if b < 0 {
		b += 2 * math.Pi
	}
	return b
}
func (b Bearing) Degrees() float64 {
	return float64(360.0 * b / (2 * math.Pi))
}

func (b Bearing) OClock() float64 {
	return b.Degrees() * 12. / 360.
}
func (b Bearing) Radians() float64 {
	return float64(b)
}

// O'clock. OR SW, etc.
func (b Bearing) String() string {
	return fmt.Sprintf("%3.f°", b.Degrees())
}
