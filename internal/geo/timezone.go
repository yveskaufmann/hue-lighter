package geo

import (
	"bufio"
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// zone1970Tab is a copy of tzdata's zone1970.tab (public domain). It lists a
// reference city with coordinates for every time zone; embedding it avoids
// depending on the file being present (e.g. in the Flatpak runtime or macOS).
//
//go:embed zone1970.tab
var zone1970Tab []byte

const localtimePath = "/etc/localtime"

func timeZoneLocation() (Location, error) {
	zone, err := systemTimeZone(os.Getenv("TZ"), localtimePath)
	if err != nil {
		return Location{}, err
	}
	return zoneLocation(zone)
}

// systemTimeZone returns the IANA name of the system time zone, e.g.
// "Europe/Berlin", from the TZ variable or the /etc/localtime symlink.
func systemTimeZone(tz string, localtime string) (string, error) {
	if tz = strings.TrimPrefix(tz, ":"); tz != "" {
		return zoneFromPath(tz), nil
	}

	target, err := filepath.EvalSymlinks(localtime)
	if err != nil {
		return "", fmt.Errorf("failed to resolve %s: %w", localtime, err)
	}
	if zone := zoneFromPath(target); zone != target {
		return zone, nil
	}
	return "", fmt.Errorf("%s does not point into a zoneinfo directory: %s", localtime, target)
}

// zoneFromPath strips everything up to the zoneinfo directory, e.g.
// "/usr/share/zoneinfo/Europe/Berlin" or macOS's
// "/var/db/timezone/zoneinfo/Europe/Berlin" become "Europe/Berlin".
func zoneFromPath(path string) string {
	if _, zone, ok := strings.Cut(path, "zoneinfo/"); ok {
		return zone
	}
	return path
}

// zoneLocation looks up the reference city coordinates of a time zone.
func zoneLocation(zone string) (Location, error) {
	scanner := bufio.NewScanner(bytes.NewReader(zone1970Tab))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) < 3 || fields[2] != zone {
			continue
		}
		latitude, longitude, err := parseISO6709(fields[1])
		if err != nil {
			return Location{}, fmt.Errorf("invalid coordinates for time zone %s: %w", zone, err)
		}
		return Location{Latitude: latitude, Longitude: longitude, Source: SourceTimeZone}, nil
	}
	return Location{}, fmt.Errorf("unknown time zone %q", zone)
}

// parseISO6709 parses coordinates in the zone.tab format: sign, degrees and
// minutes (and optionally seconds) for latitude and longitude, e.g.
// "+5230+01322" or "+404251-0740023".
func parseISO6709(coordinates string) (float64, float64, error) {
	split := strings.IndexAny(coordinates[1:], "+-") + 1
	if split <= 0 {
		return 0, 0, fmt.Errorf("malformed coordinates %q", coordinates)
	}
	latitude, err := parseDegrees(coordinates[:split], 2)
	if err != nil {
		return 0, 0, err
	}
	longitude, err := parseDegrees(coordinates[split:], 3)
	if err != nil {
		return 0, 0, err
	}
	return latitude, longitude, nil
}

func parseDegrees(value string, degreeDigits int) (float64, error) {
	sign := 1.0
	if value[0] == '-' {
		sign = -1
	}
	digits := value[1:]
	if len(digits) != degreeDigits+2 && len(digits) != degreeDigits+4 {
		return 0, fmt.Errorf("malformed coordinate %q", value)
	}

	parts := []string{digits[:degreeDigits], digits[degreeDigits : degreeDigits+2]}
	if len(digits) == degreeDigits+4 {
		parts = append(parts, digits[degreeDigits+2:])
	}

	result := 0.0
	for i, part := range parts {
		n, err := strconv.Atoi(part)
		if err != nil {
			return 0, fmt.Errorf("malformed coordinate %q: %w", value, err)
		}
		result += float64(n) / []float64{1, 60, 3600}[i]
	}
	return sign * result, nil
}
