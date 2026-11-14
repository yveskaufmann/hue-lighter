package autostart

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsAppBundleExecutable(t *testing.T) {
	assert.True(t, isAppBundleExecutable("/Applications/Hue Lighter.app/Contents/MacOS/hue-lighter"))
	assert.False(t, isAppBundleExecutable("/opt/homebrew/bin/hue-lighter"))
	assert.False(t, isAppBundleExecutable("/Users/test/Developer/hue-lighter/bin/hue-lighter"))
}

func TestLaunchAgentPlist(t *testing.T) {
	executable := "/Applications/Hue & <Lighter>.app/Contents/MacOS/hue-lighter"
	logFile := "/Users/test/Library/Logs/hue-lighter.log"

	agent := decodePlistDict(t, launchAgentPlist(executable, logFile))

	assert.Equal(t, map[string]any{
		"Label":                  LaunchAgentLabel,
		"ProgramArguments":       []any{executable},
		"RunAtLoad":              true,
		"LimitLoadToSessionType": "Aqua",
		"KeepAlive":              map[string]any{"SuccessfulExit": false},
		"ProcessType":            "Interactive",
		"EnvironmentVariables":   map[string]any{"LOG_FILE": logFile},
		"StandardErrorPath":      logFile,
	}, agent)
}

// decodePlistDict decodes the top-level dict of an XML property list, supporting
// the value types used by the LaunchAgent (string, bool, array and dict).
func decodePlistDict(t *testing.T, data []byte) map[string]any {
	t.Helper()
	decoder := xml.NewDecoder(bytes.NewReader(data))
	for {
		token, err := decoder.Token()
		require.NoError(t, err)
		if start, ok := token.(xml.StartElement); ok && start.Name.Local == "dict" {
			value, err := decodePlistValue(decoder, start)
			require.NoError(t, err)
			return value.(map[string]any)
		}
	}
}

func decodePlistValue(decoder *xml.Decoder, start xml.StartElement) (any, error) {
	switch start.Name.Local {
	case "string":
		var s string
		err := decoder.DecodeElement(&s, &start)
		return s, err
	case "true", "false":
		return start.Name.Local == "true", decoder.Skip()
	case "array", "dict":
		var items []any
		var keys []string
		for {
			token, err := decoder.Token()
			if err != nil {
				return nil, err
			}
			switch token := token.(type) {
			case xml.StartElement:
				if token.Name.Local == "key" {
					var key string
					if err := decoder.DecodeElement(&key, &token); err != nil {
						return nil, err
					}
					keys = append(keys, key)
					continue
				}
				item, err := decodePlistValue(decoder, token)
				if err != nil {
					return nil, err
				}
				items = append(items, item)
			case xml.EndElement:
				if start.Name.Local == "array" {
					return items, nil
				}
				dict := map[string]any{}
				for i, key := range keys {
					dict[key] = items[i]
				}
				return dict, nil
			}
		}
	default:
		return nil, fmt.Errorf("unsupported plist element <%s>", start.Name.Local)
	}
}
