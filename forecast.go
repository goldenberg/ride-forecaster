package main

import (
	"flag"
	"fmt"
	forecast "github.com/mlbright/forecast/v2"
	gpx "github.com/ptrv/go-gpx"
	"log"
	"strconv"
	"time"
)

var g *gpx.Gpx

var API_KEY = "806d1d0e800d3f1466ebec725982cf00"

var SanFrancisco *time.Location

// Default to Tomorrow
var start timeValue
var velocity velocityValue
var sampleInterval time.Duration

func celsiusToFahrenheit(t float64) float64 {
	return 32 + 1.8*t
}

func main() {
	start = timeValue(time.Now())
	velocity = velocityValue(NewVelocityFromMph(11))
	sampleInterval = 5 * time.Minute

	flag.Var(&start, "start", "Start time")
	flag.Var(&velocity, "velocity", "Average velocity (in mph)")
	flag.Parse()

	var fname = flag.Arg(0)
	g, err := gpx.Parse(fname)
	if err != nil {
		log.Fatalf("Error '%s' opening '%s'", err, fname)
	}

	SanFrancisco, err = time.LoadLocation("America/Los_Angeles")
	if err != nil {
		log.Fatalf("Couldn't load location %s", err)
	}

	// Load the Track from the GPX file
	track := NewTrackFromGpxWpts(g.Tracks[0].Segments[0].Points)

	// Print the start time. Parsing will fail if there is no Timestamp
	originalStart, err := time.Parse(gpx.TIMELAYOUT, g.Metadata.Timestamp)
	fmt.Printf("Original GPX start %s\n", originalStart)

	// If the user specified a time, TimeShift the track.
	var userStart = start.Get()
	if !userStart.IsZero() {
		track = track.TimeShift(userStart)
		fmt.Printf("New start: %s\n", track.times[0])
	}

	// If the user specified a velocity, model the track.
	var userVelocity = velocity.Get()
	if userVelocity != 0 {
		fmt.Println("Constant velocity: ", userVelocity.Mph())
		track = PredictTrack(track.Path(), velocity.Get(), userStart)
	}

	// Sample every n seconds, and compute a waypoint and bearing.
	for t := track.Start(); t.Before(track.End()); t = t.Add(sampleInterval) {
		wpt, bearing, err := track.WayPointAndBearingAtTime(t)
		if err != nil {
			fmt.Errorf("unable to compute intermediate waypoint at time [%s] due to ", t, err)
		}

		f, err := Forecast(wpt)
		if err != nil {
			log.Fatal(err)
		}

		windBearing := NewBearingFromDegrees(f.Currently.WindBearing)
		windAngle := (windBearing - bearing).Normalize()

		Print(wpt, f, bearing, windBearing, windAngle)
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

func Print(wpt *Waypoint, f *forecast.Forecast, bearing, windBearing, windAngle Bearing) {
	fmt.Printf("%s (%.3f, %.3f, %s): %.1f°F %.f%% %s %4.1f mph from %s at %.f o'clock.\n",
		wpt.Time.In(SanFrancisco).Format("Jan 2 03:04"),
		wpt.Lng(), wpt.Lat(), bearing,
		f.Currently.Temperature,
		f.Currently.PrecipProbability*100.,
		f.Currently.PrecipType,
		f.Currently.WindSpeed,
		windBearing,
		windAngle.OClock())
}
