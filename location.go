package main

import (
	"github.com/kelvins/sunrisesunset"
	"math"
	"time"
)

const lightWindowCoefficient float64 = 0.75

// A Location is a representation of a geographical location on the Earth's surface.
type Location struct {
	// The latitude of the geographical location in decimal format.
	Latitude float64
	// The longitude of the geographical location in decimal format.
	Longitude float64
	// The UTC offset in hours of the geographical location.
	UtcOffset float64 `yaml:"utc_offset"`
}

// An OutdoorScene is a representation of the lighting properties at a given Location and time.
type OutdoorScene struct {
	// The exact time of sunrise.
	Sunrise time.Time
	// The exact time of sunset.
	Sunset time.Time
	// The hour (0-24) when the sun crosses the meridian (mid-day).
	MeridianHour int
	// The distance in hours that the sun is from the meridian.
	DistanceFromMeridian float64
	// If the sun has already passed the meridian, this property is true.
	AfterMeridian bool
	// A lighting coefficient that uses the time of sunrise and sunset to determine how much time is spent adjusting
	// indoor lighting conditions.
	LightWindow float64
	// A lighting coefficient that determines the total length of the light window for adjusting
	// indoor lighting conditions.
	BrightnessCoefficient float64
}

// GetOutdoorScene determines the OutdoorScene for a given Location at a given point in time.
func (location *Location) GetOutdoorScene(desiredTime time.Time) (*OutdoorScene, error) {
	// calculate sunrise and sunset based on current home location
	sunrise, sunset, err := sunrisesunset.GetSunriseSunset(location.Latitude, location.Longitude, location.UtcOffset, desiredTime)
	if err != nil {
		return nil, err
	}

	outdoorScene := OutdoorScene{}
	outdoorScene.Sunrise = sunrise
	outdoorScene.Sunset = sunset

	// A few coefficients to tune the window in which the lighting defaults change
	outdoorScene.MeridianHour = (sunrise.Hour() + sunset.Hour()) / 2
	outdoorScene.LightWindow = float64(sunset.Hour() - sunrise.Hour()) / lightWindowCoefficient
	outdoorScene.AfterMeridian = desiredTime.Hour() - outdoorScene.MeridianHour > 0
	outdoorScene.DistanceFromMeridian = math.Abs(float64(desiredTime.Hour() - outdoorScene.MeridianHour))

	// Determine the length of the light window
	outdoorScene.BrightnessCoefficient = outdoorScene.LightWindow - outdoorScene.DistanceFromMeridian
	if outdoorScene.BrightnessCoefficient < 0 {
		outdoorScene.BrightnessCoefficient = 0
	}

	return &outdoorScene, nil
}