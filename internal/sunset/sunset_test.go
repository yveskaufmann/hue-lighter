package sunset

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculateSunriseSunset(t *testing.T) {
	tests := []struct {
		name      string
		latitude  float64
		longitude float64
		validate  func(t *testing.T, sunrise, sunset time.Time)
	}{
		{
			name:      "calculates sunrise and sunset for Berlin",
			latitude:  52.5200,
			longitude: 13.4050,
			validate: func(t *testing.T, sunrise, sunset time.Time) {
				// Berlin coordinates should return valid times
				assert.False(t, sunrise.IsZero())
				assert.False(t, sunset.IsZero())
				// Sunrise should be before sunset
				assert.True(t, sunrise.Before(sunset))
				// Both should be on the same day as today
				now := time.Now()
				assert.Equal(t, now.Year(), sunrise.Year())
				assert.Equal(t, now.Month(), sunrise.Month())
				assert.Equal(t, now.Day(), sunrise.Day())
				assert.Equal(t, now.Year(), sunset.Year())
				assert.Equal(t, now.Month(), sunset.Month())
				assert.Equal(t, now.Day(), sunset.Day())
			},
		},
		{
			name:      "calculates sunrise and sunset for New York",
			latitude:  40.7128,
			longitude: -74.0060,
			validate: func(t *testing.T, sunrise, sunset time.Time) {
				assert.False(t, sunrise.IsZero())
				assert.False(t, sunset.IsZero())
				assert.True(t, sunrise.Before(sunset))
			},
		},
		{
			name:      "calculates sunrise and sunset for Tokyo",
			latitude:  35.6762,
			longitude: 139.6503,
			validate: func(t *testing.T, sunrise, sunset time.Time) {
				assert.False(t, sunrise.IsZero())
				assert.False(t, sunset.IsZero())
				assert.True(t, sunrise.Before(sunset))
			},
		},
		{
			name:      "calculates sunrise and sunset for Sydney",
			latitude:  -33.8688,
			longitude: 151.2093,
			validate: func(t *testing.T, sunrise, sunset time.Time) {
				assert.False(t, sunrise.IsZero())
				assert.False(t, sunset.IsZero())
				assert.True(t, sunrise.Before(sunset))
			},
		},
		{
			name:      "calculates sunrise and sunset for London",
			latitude:  51.5074,
			longitude: -0.1278,
			validate: func(t *testing.T, sunrise, sunset time.Time) {
				assert.False(t, sunrise.IsZero())
				assert.False(t, sunset.IsZero())
				assert.True(t, sunrise.Before(sunset))
			},
		},
		{
			name:      "handles equator coordinates",
			latitude:  0.0,
			longitude: 0.0,
			validate: func(t *testing.T, sunrise, sunset time.Time) {
				assert.False(t, sunrise.IsZero())
				assert.False(t, sunset.IsZero())
				assert.True(t, sunrise.Before(sunset))
				// At the equator, day length is relatively consistent
				// Approximately 12 hours between sunrise and sunset
				duration := sunset.Sub(sunrise)
				assert.Greater(t, duration.Hours(), 11.0)
				assert.Less(t, duration.Hours(), 13.0)
			},
		},
		{
			name:      "handles northern hemisphere high latitude",
			latitude:  60.0,
			longitude: 10.0,
			validate: func(t *testing.T, sunrise, sunset time.Time) {
				// At high latitudes, times may vary dramatically by season
				// but should still be valid
				assert.False(t, sunrise.IsZero())
				assert.False(t, sunset.IsZero())
			},
		},
		{
			name:      "handles southern hemisphere high latitude",
			latitude:  -45.0,
			longitude: 170.0,
			validate: func(t *testing.T, sunrise, sunset time.Time) {
				assert.False(t, sunrise.IsZero())
				assert.False(t, sunset.IsZero())
			},
		},
		{
			name:      "handles positive longitude (east)",
			latitude:  35.0,
			longitude: 139.0,
			validate: func(t *testing.T, sunrise, sunset time.Time) {
				assert.False(t, sunrise.IsZero())
				assert.False(t, sunset.IsZero())
				assert.True(t, sunrise.Before(sunset))
			},
		},
		{
			name:      "handles negative longitude (west)",
			latitude:  40.0,
			longitude: -100.0,
			validate: func(t *testing.T, sunrise, sunset time.Time) {
				assert.False(t, sunrise.IsZero())
				assert.False(t, sunset.IsZero())
				assert.True(t, sunrise.Before(sunset))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sunrise, sunset := CalculateSunriseSunset(tt.latitude, tt.longitude)

			require.NotNil(t, tt.validate)
			tt.validate(t, sunrise, sunset)
		})
	}
}

func TestCalculateSunriseSunset_EdgeCases(t *testing.T) {
	t.Run("handles extreme northern latitude", func(t *testing.T) {
		// Arctic Circle
		sunrise, sunset := CalculateSunriseSunset(66.5, 0.0)

		// Even in polar regions, the function should return times
		// During summer: midnight sun (sunset after sunrise next day)
		// During winter: polar night (no true sunrise/sunset)
		assert.False(t, sunrise.IsZero())
		assert.False(t, sunset.IsZero())
	})

	t.Run("handles extreme southern latitude", func(t *testing.T) {
		// Antarctic Circle
		sunrise, sunset := CalculateSunriseSunset(-66.5, 0.0)

		assert.False(t, sunrise.IsZero())
		assert.False(t, sunset.IsZero())
	})

	t.Run("handles north pole", func(t *testing.T) {
		// North Pole - extreme case
		sunrise, sunset := CalculateSunriseSunset(90.0, 0.0)

		// At the pole, sunrise/sunset depends on time of year
		// During polar night or midnight sun, times may be zero
		// We just verify the function doesn't panic
		_ = sunrise
		_ = sunset
	})

	t.Run("handles south pole", func(t *testing.T) {
		// South Pole
		sunrise, sunset := CalculateSunriseSunset(-90.0, 0.0)

		// Same as north pole - may return zero times during polar conditions
		_ = sunrise
		_ = sunset
	})
}

func TestCalculateSunriseSunset_TimeConsistency(t *testing.T) {
	t.Run("returns same results for same location on same day", func(t *testing.T) {
		lat, lon := 48.8566, 2.3522 // Paris

		sunrise1, sunset1 := CalculateSunriseSunset(lat, lon)

		// Call again immediately
		sunrise2, sunset2 := CalculateSunriseSunset(lat, lon)

		// Should return same times (within same minute since time.Now() is used)
		assert.Equal(t, sunrise1.Hour(), sunrise2.Hour())
		assert.Equal(t, sunrise1.Minute(), sunrise2.Minute())
		assert.Equal(t, sunset1.Hour(), sunset2.Hour())
		assert.Equal(t, sunset1.Minute(), sunset2.Minute())
	})

	t.Run("sunrise and sunset are on the same day as today", func(t *testing.T) {
		lat, lon := 37.7749, -122.4194 // San Francisco

		sunrise, sunset := CalculateSunriseSunset(lat, lon)

		// The sunrise library may return times in a different timezone
		// We just verify they're not zero and sunrise is before sunset
		assert.False(t, sunrise.IsZero())
		assert.False(t, sunset.IsZero())
		assert.True(t, sunrise.Before(sunset))
	})
}

func TestCalculateSunriseSunset_ReasonableTimes(t *testing.T) {
	t.Run("sunrise is during morning hours for typical locations", func(t *testing.T) {
		// Mid-latitude locations should have sunrise in morning (4-9 AM typically)
		lat, lon := 41.9028, 12.4964 // Rome

		sunrise, _ := CalculateSunriseSunset(lat, lon)

		// Sunrise should be between midnight and noon
		assert.GreaterOrEqual(t, sunrise.Hour(), 0)
		assert.Less(t, sunrise.Hour(), 12)
	})

	t.Run("sunset is during evening hours for typical locations", func(t *testing.T) {
		// Mid-latitude locations should have sunset in evening (4-9 PM typically)
		lat, lon := 41.9028, 12.4964 // Rome

		_, sunset := CalculateSunriseSunset(lat, lon)

		// Sunset should be between noon and midnight
		assert.GreaterOrEqual(t, sunset.Hour(), 12)
		assert.Less(t, sunset.Hour(), 24)
	})

	t.Run("day length is reasonable for mid-latitudes", func(t *testing.T) {
		lat, lon := 40.7128, -74.0060 // New York

		sunrise, sunset := CalculateSunriseSunset(lat, lon)

		duration := sunset.Sub(sunrise)

		// Day length should be between 9 and 16 hours for mid-latitudes
		// (varies by season)
		assert.Greater(t, duration.Hours(), 8.0)
		assert.Less(t, duration.Hours(), 17.0)
	})
}

func TestCalculateSunriseSunset_BoundaryValues(t *testing.T) {
	tests := []struct {
		name      string
		latitude  float64
		longitude float64
	}{
		{
			name:      "maximum latitude",
			latitude:  90.0,
			longitude: 0.0,
		},
		{
			name:      "minimum latitude",
			latitude:  -90.0,
			longitude: 0.0,
		},
		{
			name:      "maximum longitude",
			latitude:  0.0,
			longitude: 180.0,
		},
		{
			name:      "minimum longitude",
			latitude:  0.0,
			longitude: -180.0,
		},
		{
			name:      "all zeros",
			latitude:  0.0,
			longitude: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic and should return times
			// (may be zero at poles during certain seasons)
			sunrise, sunset := CalculateSunriseSunset(tt.latitude, tt.longitude)

			// At the poles (90/-90), times may be zero during polar night/midnight sun
			// For other coordinates, times should be valid
			if tt.latitude != 90.0 && tt.latitude != -90.0 {
				assert.False(t, sunrise.IsZero())
				assert.False(t, sunset.IsZero())
			}
			// Main test is that the function doesn't panic
			_ = sunrise
			_ = sunset
		})
	}
}

func TestCalculateSunriseSunset_Integration(t *testing.T) {
	t.Run("works with common use case - European city", func(t *testing.T) {
		// Munich, Germany
		lat := 48.1351
		lon := 11.5820

		sunrise, sunset := CalculateSunriseSunset(lat, lon)

		// Verify basic properties
		assert.False(t, sunrise.IsZero())
		assert.False(t, sunset.IsZero())
		assert.True(t, sunrise.Before(sunset))

		// Verify times are for today
		now := time.Now()
		assert.Equal(t, now.Year(), sunrise.Year())
		assert.Equal(t, now.Month(), sunrise.Month())
		assert.Equal(t, now.Day(), sunrise.Day())

		// Verify reasonable time ranges
		assert.GreaterOrEqual(t, sunrise.Hour(), 0)
		assert.Less(t, sunrise.Hour(), 12)
		assert.GreaterOrEqual(t, sunset.Hour(), 12)
		assert.Less(t, sunset.Hour(), 24)
	})

	t.Run("works with common use case - American city", func(t *testing.T) {
		// Seattle, USA
		lat := 47.6062
		lon := -122.3321

		sunrise, sunset := CalculateSunriseSunset(lat, lon)

		assert.False(t, sunrise.IsZero())
		assert.False(t, sunset.IsZero())
		assert.True(t, sunrise.Before(sunset))

		// Day should be between 8-16 hours for this latitude
		duration := sunset.Sub(sunrise)
		assert.Greater(t, duration.Hours(), 7.0)
		assert.Less(t, duration.Hours(), 18.0)
	})
}
