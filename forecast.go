package main

import (
	"flag"
	"fmt"
	"github.com/bradfitz/gomemcache/memcache"
	forecast "github.com/mlbright/forecast/v2"
	gpx "github.com/ptrv/go-gpx"
	"log"
	"strconv"
	"time"
	"encoding/json"
)

var g *gpx.Gpx

var API_KEY = "806d1d0e800d3f1466ebec725982cf00"

var SanFrancisco *time.Location

var start timeValue
var velocity velocityValue
var sampleInterval time.Duration
var server bool

func main() {
	start = timeValue(time.Now())
	velocity = velocityValue(NewVelocityFromMph(11))
	sampleInterval = 5 * time.Minute
	server = false

	flag.BoolVar(&server, "server", false, "Run in server mode.")
	// TODO: default to tomorrow at 8am
	flag.Var(&start, "start", "Start time")
	flag.Var(&velocity, "velocity", "Average velocity (in mph)")
	flag.Parse()

	if server {
		startServer()
	} else {
		var fname = flag.Arg(0)
		var startTime = start.Get()
		var userVelocity = velocity.Get()

		track, _ := ReadTrack(fname)
		track = ModelTrack(track, startTime, userVelocity)

		data := make([]ForecastedLocation, 0)
		for f := range ForecastTrack(track, sampleInterval) {
			f.Print()
			data = append(data, *f)
		}
		return
	}

}

// ReadTrack reads a Track from a GPX 1.1 file
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
	t = NewTrackFromGpxWpts(g.Tracks[0].Segments[0].Points)

	// Print the start time. Parsing will fail if there is no Timestamp
	// originalStart, _ := time.Parse(gpx.TIMELAYOUT, g.Metadata.Timestamp)

	return t, err
}

// ModelTrack translates a Track in time and rescales it to have a different constant velocity.
// If startTime is zero, the track won't be translated, and if the velocity is zero the track
// won't be rescaled.
func ModelTrack(track *Track, startTime time.Time, velocity Velocity) (out *Track) {
	// If the user specified a velocity, rescale the track.
	if velocity != 0 {
		out = PredictTrack(track.Path(), velocity, startTime)
	}

	// If the user specified a time, TimeShift the track.
	if !startTime.IsZero() {
		out = out.TimeShift(startTime)
	}
	return out
}

// ForecastTrack samples the track at a time interval, interpolating as necessary
// computes the wind bearings, and queries for a forecast.
func ForecastTrack(track *Track, sampleInterval time.Duration) (out chan *ForecastedLocation) {

	out = make(chan *ForecastedLocation, 0)
	// Sample every n seconds, and compute a waypoint and bearing.
	go func() {
		for t := track.Start(); t.Before(track.End()); t = t.Add(sampleInterval) {
			wpt, bearing, err := track.Interpolate(t)
			if err != nil {
				fmt.Errorf("unable to compute intermediate waypoint at time [%s] due to ", t, err)
			}

			useCache := true
			var f *forecast.Forecast
			if useCache {
				f, err = lookupCache(wpt)
			} else {
				f, err = ForecastWaypoint(wpt)
			}
			if err != nil {
				log.Fatal(err)
			}

			windBearing := NewBearingFromDegrees(f.Currently.WindBearing)
			windAngle := (windBearing - bearing).Normalize()

			pt := &ForecastedLocation{f, wpt, bearing, windAngle}
			out <- pt
		}
		close(out)
	}()
	return out
}

func lookupCache(wpt *Waypoint) (f *forecast.Forecast, err error) {
	mc := memcache.New("localhost:11211")

	cacheTime := wpt.Time.Round(time.Duration(time.Minute * 10))
	cacheKey := fmt.Sprintf("%.4f,%.4f,%v", wpt.Lng(), wpt.Lat(), cacheTime.Format("01/02/2006.15:04"))
	it, err := mc.Get(cacheKey)

	if err == memcache.ErrCacheMiss {
		f, err := ForecastWaypoint(wpt)
		if err != nil {
			return nil, err
		}
		fmt.Println("cache miss for ", cacheKey)
		val, err := json.Marshal(f)
		if err != nil {
			return nil, err
		}
		it := &memcache.Item{Key: cacheKey, Value: val}
		err = mc.Set(it)
		if err != nil {
			return nil, err
		}
		return f, nil
	} else if err != nil {
		return nil, err
	} else {
		fmt.Println("cache hit for ", cacheKey)
		err = json.Unmarshal(it.Value, &f)
		if err != nil {
			return nil, err
		}
		return f, nil
	}
}

func ForecastWaypoint(wpt *Waypoint) (f *forecast.Forecast, err error) {
	f, err = forecast.Get(API_KEY,
		fmt.Sprintf("%.4f", wpt.Lng()),
		fmt.Sprintf("%.4f", wpt.Lat()),
		strconv.FormatInt(wpt.Time.Unix(), 10),
		forecast.US)
	return
}

type ForecastedLocation struct {
	Forecast  *forecast.Forecast `json:"forecast"`
	Waypoint  *Waypoint          `json:"waypoint"`
	Heading   Bearing            `json:"heading"`
	WindAngle Bearing            `json:"windAngle"`
}

func (d *ForecastedLocation) Print() {
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
