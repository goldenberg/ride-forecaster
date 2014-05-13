package main

import (
	"flag"
	"fmt"
	"strconv"
	// rideforecaster "github.com/goldenberg/rideforecaster"
	forecast "github.com/mlbright/forecast/v2"
	gpx "github.com/ptrv/go-gpx"
	"log"
	"time"
)

var g *gpx.Gpx

var API_KEY = "806d1d0e800d3f1466ebec725982cf00"

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
		f, err := Forecast(wpt)
		if err != nil {
			log.Fatal(err)
		}

		Print(wpt, f)
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

func Print(wpt *Waypoint, f *forecast.Forecast) {
	fmt.Printf("(%.4f, %.4f) %s: %.1f° %.1f mph %.0f \n",
		wpt.Lat(), wpt.Lng(), wpt.Time.Format(gpx.TIMELAYOUT),
		celsiusToFahrenheit(f.Currently.Temperature),
		f.Currently.WindSpeed,
		f.Currently.WindBearing)
}
