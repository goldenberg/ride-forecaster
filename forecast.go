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

	data := make([]DataPoint, 0)
	for f := range ForecastTrack(track, sampleInterval) {
		f.Print()
		data = append(data, *f)
	}
}

func ForecastTrack(track *Track, sampleInterval time.Duration) (out chan *DataPoint) {
	out = make(chan *DataPoint, 0)
	// Sample every n seconds, and compute a waypoint and bearing.
	go func() {
		for t := track.Start(); t.Before(track.End()); t = t.Add(sampleInterval) {
			wpt, bearing, err := track.Interpolate(t)
			if err != nil {
				fmt.Errorf("unable to compute intermediate waypoint at time [%s] due to ", t, err)
			}

			f, err := ForecastWaypoint(wpt)
			if err != nil {
				log.Fatal(err)
			}

			windBearing := NewBearingFromDegrees(f.Currently.WindBearing)
			windAngle := (windBearing - bearing).Normalize()

			pt := &DataPoint{f, wpt, bearing, windAngle}
			out <- pt
		}
		close(out)
	}()
	return out
}

func ForecastWaypoint(wpt *Waypoint) (f *forecast.Forecast, err error) {
	f, err = forecast.Get(API_KEY,
		fmt.Sprintf("%.4f", wpt.Lng()),
		fmt.Sprintf("%.4f", wpt.Lat()),
		strconv.FormatInt(wpt.Time.Unix(), 10),
		forecast.US)
	return
}

type DataPoint struct {
	f         *forecast.Forecast
	wpt       *Waypoint
	bearing   Bearing
	windAngle Bearing
}

func (d *DataPoint) Print() {
	fmt.Printf("%s (%.3f, %.3f, %s): %.1f°F %.f%% %s at %.3f in/hr  Wind: %2.1f mph from %s at %.f o'clock.\n",
		d.wpt.Time.In(SanFrancisco).Format("Jan 2 03:04"),
		d.wpt.Lng(), d.wpt.Lat(), d.bearing,
		d.f.Currently.Temperature,
		d.f.Currently.PrecipProbability*100.,
		d.f.Currently.PrecipType,
		d.f.Currently.PrecipIntensity,
		d.f.Currently.WindSpeed,
		NewBearingFromDegrees(d.f.Currently.WindBearing),
		d.windAngle.OClock())
}
