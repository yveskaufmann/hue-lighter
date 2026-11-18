package buildinfo

import (
	"fmt"
	"strings"
)

var (
	version = "dev"
	commit  = "none"
	buildDate = "unknown"
)

func Version() string {
	return strings.TrimSpace(version)
}

func Commit() string {
	return strings.TrimSpace(commit)
}

func BuildDate() string {
	return strings.TrimSpace(buildDate)
}

func Format() string {

	lines := []string{
		fmt.Sprintf("%-11s%s", "Version:",  version),
		fmt.Sprintf("%-11s%s", "Commit:",  commit),
		fmt.Sprintf("%-11s%s", "BuildDate:",  buildDate),
	}

	return strings.Join(lines, "\n")
}
