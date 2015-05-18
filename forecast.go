package main

import (
	"encoding/json"
	"flag"
	"fmt"
	forecast "github.com/mlbright/forecast/v2"
	gpx "github.com/ptrv/go-gpx"
	"log"
	"net/http"
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
var server bool

func celsiusToFahrenheit(t float64) float64 {
	return 32 + 1.8*t
}

func main() {
	start = timeValue(time.Now())
	velocity = velocityValue(NewVelocityFromMph(11))
	sampleInterval = 50 * time.Minute
	server = false

	flag.BoolVar(&server, "server", false, "Run in server mode.")
	flag.Var(&start, "start", "Start time")
	flag.Var(&velocity, "velocity", "Average velocity (in mph)")
	flag.Parse()

	if server {
		http.HandleFunc("/forecast", handler)
		http.ListenAndServe(":8080", nil)

	} else {
		var fname = flag.Arg(0)
		var startTime = start.Get()
		var userVelocity = velocity.Get()

		track, _ := ReadTrack(fname)
		track = ModelTrack(track, startTime, userVelocity)

		data := make([]DataPoint, 0)
		for f := range ForecastTrack(track, sampleInterval) {
			f.Print()
			data = append(data, *f)
		}
		return
	}

}

func handler(w http.ResponseWriter, r *http.Request) {
	// fmt.Fprintf(w, "handling %s", r)

	var startTime = time.Now()
	var velocity = NewVelocityFromMph(10)
	track, _ := ReadTrack("data/gpx_11/bofax_alpine11.gpx")

	fmt.Println("Track has %i points", track.Path().Length())
	track = ModelTrack(track, startTime, velocity)
	data := make([]DataPoint, 0)
	for d := range ForecastTrack(track, sampleInterval) {
		d.Print()
		data = append(data, *d)
	}

	b, err := json.Marshal(data)
	if err != nil {
		fmt.Fprintf(w, "Error: %s", err)
	}
	w.Write(b)
}

func ReadTrack(fname string) (t *Track, err error) {
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
	originalStart, _ := time.Parse(gpx.TIMELAYOUT, g.Metadata.Timestamp)
	fmt.Printf("Original GPX start %s\n", originalStart)

	return track, err
}

func ModelTrack(track *Track, startTime time.Time, velocity Velocity) (out *Track) {
	// If the user specified a velocity, model the track.
	if velocity != 0 {
		fmt.Println("Constant velocity: ", velocity.Mph())
		out = PredictTrack(track.Path(), velocity, startTime)
	}

	// If the user specified a time, TimeShift the track.
	if !startTime.IsZero() {
		out = out.TimeShift(startTime)
		fmt.Printf("New start: %s\n", track.times[0])
	}
	return out
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
	Forecast  *forecast.Forecast `json:"forecast"`
	Waypoint  *Waypoint          `json:"waypoint"`
	Heading   Bearing            `json:"heading"`
	WindAngle Bearing            `json:"windAngle"`
}

func (d *DataPoint) Print() {
	fmt.Printf("%s (%.3f, %.3f, %s): %.1f°F %.f%% %s at %.3f in/hr  Wind: %2.1f mph from %s at %.f o'clock.\n",
		d.Waypoint.Time.In(SanFrancisco).Format("Jan 2 03:04"),
		d.Waypoint.Lng(), d.Waypoint.Lat(), d.Heading,
		d.Forecast.Currently.Temperature,
		d.Forecast.Currently.PrecipProbability*100.,
		d.Forecast.Currently.PrecipType,
		d.Forecast.Currently.PrecipIntensity,
		d.Forecast.Currently.WindSpeed,
		NewBearingFromDegrees(d.Forecast.Currently.WindBearing),
		d.WindAngle.OClock())
}
