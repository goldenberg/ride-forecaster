package cli

import (
	"flag"
	"time"

	"github.com/goldenberg/ride-forecaster/pkg/types"
	gpx "github.com/ptrv/go-gpx"
)

var g *gpx.Gpx

var apiKey = "806d1d0e800d3f1466ebec725982cf00"
var sanFrancisco *time.Location

var start timeValue
var velocity velocityValue
var sampleInterval time.Duration
var server bool

func main() {
	start = timeValue(time.Now())
	velocity = velocityValue(types.NewVelocityFromMph(11))
	sampleInterval = 5 * time.Minute
	server = false

	// TODO: default to tomorrow at 8am
	flag.Var(&start, "start", "Start time")
	flag.Var(&velocity, "velocity", "Average velocity (in mph)")
	flag.Parse()

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
