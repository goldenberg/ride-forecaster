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

type Bearing float64

func NewBearing(r float64) Bearing {
	return Bearing(r).Normalize()
}

func NewBearingFromDegrees(d float64) Bearing {
	return Bearing(deg2Rad(d)).Normalize()
}

func (b Bearing) Normalize() Bearing {
	if b < 0 {
		b += 2 * math.Pi
	}
	return b
}
func (b Bearing) Degrees() float64 {
	return float64(360.0 * b / (2 * math.Pi))
}

func (b Bearing) OClock() float64 {
	return b.Degrees() * 12. / 360.
}
func (b Bearing) Radians() float64 {
	return float64(b)
}

// O'clock. OR SW, etc.
func (b Bearing) String() string {
	return fmt.Sprintf("%3.f°", b.Degrees())
}
