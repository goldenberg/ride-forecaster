package main

import (
	"fmt"
	geo "github.com/paulmach/go.geo"
	"math"
	"time"
)

type Waypoint struct {
	*geo.Point
	Time time.Time
}

// Velocity is measured in m/s
type Velocity float64

func NewVelocityFromMph(mph float64) Velocity {
	return Velocity(mph * 1609.34 / 3600.)

}
func (v Velocity) Mph() float64 {
	return float64(v/1609.34) * 3600
}

func (v Velocity) Ms() float64 {
	return float64(v)
}

// Bearing is measured in radians.
type Bearing float64

// NewBearing creates a Bearing at r radians.
func NewBearing(r float64) Bearing {
	return Bearing(r).Normalize()
}

// NewBearingFromDegrees creates a Bearing at r radians.
func NewBearingFromDegrees(d float64) Bearing {
	return Bearing(deg2Rad(d)).Normalize()
}

func (b Bearing) Normalize() Bearing {
	if b < 0 {
		b += 2 * math.Pi
	}
	return b
}

// Degrees returns the Bearing in degrees from due North.
func (b Bearing) Degrees() float64 {
	return float64(360.0 * b / (2 * math.Pi))
}

// OClock returns the Bearing as represented on a 12 hour clock face.
// e.g. 270 degrees is 9 o'clock
func (b Bearing) OClock() float64 {
	return b.Degrees() * 12. / 360.
}

// Radians returns the Bearing in radians between 0 and 2π
func (b Bearing) Radians() float64 {
	return float64(b)
}

func (b Bearing) String() string {
	return fmt.Sprintf("%3.f°", b.Degrees())
}

func deg2Rad(d float64) float64 {
	return d * math.Pi / 180.0
}

func rad2Deg(r float64) float64 {
	return 180.0 * r / math.Pi
}
