package main

import (
	"flag"
	"fmt"
	"math"
	"strconv"
	// rideforecaster "github.com/goldenberg/rideforecaster"
	forecast "github.com/mlbright/forecast/v2"
	gpx "github.com/ptrv/go-gpx"
	"log"
	"time"
)

var g *gpx.Gpx

var API_KEY = "806d1d0e800d3f1466ebec725982cf00"

var SanFrancisco *time.Location

func celsiusToFahrenheit(t float64) float64 {
	return 32 + 1.8*t
}

func main() {
	flag.Parse()
	var fname = flag.Arg(0)
	g, err := gpx.Parse(fname)
	if err != nil {
		log.Fatalf("Error '%s' opening '%s'", err, fname)
	}

	SanFrancisco, err = time.LoadLocation("America/Los_Angeles")
	if err != nil {
		log.Fatal("Couldn't load location", err)
	}

	// Print the start time
	start, err := time.Parse(gpx.TIMELAYOUT, g.Metadata.Timestamp)
	if err != nil {
		log.Fatalf("Error '%s' parsing timestamp '%s'", err, g.Metadata.Timestamp)
	}
	fmt.Printf("GPX start %s\n", start)

	// Print weather at every nth point
	track := NewTrackFromGpxWpts(g.Tracks[0].Segments[0].Points)

	for i := 0; i < track.Length(); i++ {
		if i%30 != 0 {
			continue
		}
		wpt := track.Waypoint(i)
		next := track.Path.GetAt(i + 1)

		f, err := Forecast(wpt)
		if err != nil {
			log.Fatal(err)
		}

		bearing := NewBearingFromDegrees(wpt.BearingTo(next))
		windBearing := NewBearingFromDegrees(f.Currently.WindBearing)

		windAngle := (windBearing - bearing).Normalize()
		effectiveHeadwind := math.Cos(float64(windAngle)) * f.Currently.WindSpeed

		Print(wpt, f, bearing, windBearing, windAngle, effectiveHeadwind)
	}
}

func Forecast(wpt *Waypoint) (f *forecast.Forecast, err error) {
	f, err = forecast.Get(API_KEY,
		fmt.Sprintf("%.4f", wpt.Lng()),
		fmt.Sprintf("%.4f", wpt.Lat()),
		strconv.FormatInt(wpt.Time.Unix(), 10),
		forecast.US)
	return
}

func Print(wpt *Waypoint, f *forecast.Forecast, bearing, windBearing, windAngle Bearing, effectiveHeadwind float64) {
	fmt.Printf("%s (%.3f, %.3f, %s): %.1f°F %4.1f mph at %s.   Effective: %5.1f mph at %s\n",
		wpt.Time.In(SanFrancisco).Format("03:04"), wpt.Lng(), wpt.Lat(), bearing,
		f.Currently.Temperature,
		f.Currently.WindSpeed,
		windBearing,
		effectiveHeadwind,
		windAngle)
}
