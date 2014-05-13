package main

import (
	geo "github.com/paulmach/go.geo"
	gpx "github.com/ptrv/go-gpx"
	"log"
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
