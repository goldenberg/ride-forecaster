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

var increment = 5 * time.Minute

func main() {
	// lat := "43.6595"
	// long := "-79.3433"

	flag.Parse()
	var fname = flag.Arg(0)
	g, err := gpx.Parse(fname)
	if err != nil {
		log.Fatalf("Error '%s' opening '%s'", err, fname)
	}
	// fmt.Printf("GPX:", g)
	fmt.Printf("GPX timestamp %s", g.Metadata.Timestamp)
	fmt.Printf("GPX bounds %s", g.Bounds())

	var segment = g.Tracks[0].Segments[0]
	// fmt.Printf("Gpx track segments %s", segment)

	for _, pt := range segment.Points {
		f, err := forecast.Get(API_KEY,
			fmt.Sprintf("%.4f", pt.Lat),
			fmt.Sprintf("%.4f", pt.Lon),
			pt.Timestamp, forecast.CA)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("(%.4f, %.4f) %s: %s \n",
			pt.Lat, pt.Lon,
			pt.Timestamp,
			celsiusToFahrenheit(f.Currently.Temperature))
		// fmt.Printf("%s: %s\n", f.Timezone, f.Currently.Summary)
		// fmt.Printf("humidity: %.2f\n", f.Currently.Humidity)
		// fmt.Printf("temperature: %.2f Celsius\n", f.Currently.Temperature)
		// fmt.Printf("wind speed: %.2f\n", f.Currently.WindSpeed)
	}

	// fmt.Printf("Ride starting at %s", g.Metadata.Timestamp)
}
