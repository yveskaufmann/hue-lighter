package logging

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOutput(t *testing.T) {
	t.Run("defaults to stderr", func(t *testing.T) {
		t.Setenv("LOG_FILE", "")

		assert.Same(t, os.Stderr, newOutput())
	})

	t.Run("appends to LOG_FILE and creates missing directories", func(t *testing.T) {
		logFile := filepath.Join(t.TempDir(), "Logs", "hue-lighter.log")
		require.NoError(t, os.MkdirAll(filepath.Dir(logFile), 0o755))
		require.NoError(t, os.WriteFile(logFile, []byte("previous run\n"), 0o644))
		t.Setenv("LOG_FILE", logFile)
		t.Setenv("LOG_FORMAT", "text")

		logger := NewLogger()
		file, ok := logger.Logger.Out.(*os.File)
		require.True(t, ok, "expected the logger to write to a file")
		t.Cleanup(func() { file.Close() })

		logger.Info("hello from the logger")

		content, err := os.ReadFile(logFile)
		require.NoError(t, err)
		assert.Contains(t, string(content), "previous run\n")
		assert.Contains(t, string(content), "hello from the logger")
	})

	t.Run("expands a leading ~ to the home directory", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)
		t.Setenv("LOG_FILE", "~/Library/Logs/hue-lighter.log")

		output := newOutput()
		file, ok := output.(*os.File)
		require.True(t, ok, "expected a log file")
		t.Cleanup(func() { file.Close() })

		assert.Equal(t, filepath.Join(home, "Library", "Logs", "hue-lighter.log"), file.Name())
		assert.FileExists(t, file.Name())
	})

	t.Run("falls back to stderr if LOG_FILE cannot be opened", func(t *testing.T) {
		notADirectory := filepath.Join(t.TempDir(), "file")
		require.NoError(t, os.WriteFile(notADirectory, nil, 0o644))
		t.Setenv("LOG_FILE", filepath.Join(notADirectory, "hue-lighter.log"))

		assert.Same(t, os.Stderr, newOutput())
	})
}

func TestExpandHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	for input, want := range map[string]string{
		"~":                 home,
		"~/hue-lighter.log": filepath.Join(home, "hue-lighter.log"),
		"/var/log/hue.log":  "/var/log/hue.log",
		"relative/hue.log":  "relative/hue.log",
		"~other/hue.log":    "~other/hue.log",
	} {
		got, err := expandHome(input)
		require.NoError(t, err)
		assert.Equal(t, want, got, "input %q", input)
	}
}
