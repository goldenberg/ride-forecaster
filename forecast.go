package main

import (
	"flag"
	"fmt"
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
	var track = g.Tracks[0]
	for i, pt := range track.Segments[0].Points {
		if i%25 != 0 {
			continue
		}
		f, err := Forecast(pt)
		if err != nil {
			log.Fatal(err)
		}

		Print(pt, f)
	}
}

func Forecast(pt gpx.GpxWpt) (f *forecast.Forecast, err error) {
	f, err = forecast.Get(API_KEY,
		fmt.Sprintf("%.4f", pt.Lat),
		fmt.Sprintf("%.4f", pt.Lon),
		pt.Timestamp, forecast.CA)
	return
}

func Print(pt gpx.GpxWpt, f *forecast.Forecast) {
	fmt.Printf("(%.4f, %.4f) %s: %.1f %.1f mph \n",
		pt.Lat, pt.Lon, pt.Timestamp,
		celsiusToFahrenheit(f.Currently.Temperature),
		f.Currently.Windspeed)
}
