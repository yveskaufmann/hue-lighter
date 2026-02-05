package sunset

import (
	"time"

	"github.com/nathan-osman/go-sunrise"
)

func CalculateSunriseSunset(latitude float64, longitude float64) (time.Time, time.Time) {
	now := time.Now()

	sunriseTime, sunsetTime := sunrise.SunriseSunset(
		latitude,
		longitude,
		now.Year(),
		now.Month(),
		now.Day(),
	)

	return sunriseTime, sunsetTime
}
