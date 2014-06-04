package main

import (
	"fmt"
	geo "github.com/paulmach/go.geo"
	gpx "github.com/ptrv/go-gpx"
	"log"
	"math"
	"time"
)

type Waypoint struct {
	*geo.Point
	Time time.Time
}

type Track struct {
	*geo.Path
	times []time.Time
}

func (t *Track) Waypoint(i int) *Waypoint {
	return &Waypoint{t.GetAt(i), t.times[i]}
}

func NewTrackFromGpxWpts(wpts []gpx.GpxWpt) (track *Track) {
	times := make([]time.Time, len(wpts), len(wpts))
	points := make([]geo.Point, len(wpts), len(wpts))

	for i, wpt := range wpts {
		points[i] = *geo.NewPoint(wpt.Lat, wpt.Lon)
		t, err := time.Parse(gpx.TIMELAYOUT, wpt.Timestamp)
		if err != nil {
			log.Fatalf("Error '%s' parsing timestamp '%s'", err, wpt.Timestamp)
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
	return float64(180.0 * b / math.Pi)
}

func (b Bearing) Radians() float64 {
	return float64(b)
}

// O'clock. OR SW, etc.
func (b Bearing) String() string {
	return fmt.Sprintf("%3.f°", b.Degrees())
}
