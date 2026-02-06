package logging

import (
	"os"
	"testing"

	"com.github.yveskaufmann/hue-lighter/internal/testutils"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLogger(t *testing.T) {
	t.Run("creates new logger with default settings", func(t *testing.T) {
		// Clear any environment variables
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("LOG_FORMAT")

		logger := NewLogger()

		require.NotNil(t, logger)
		require.NotNil(t, logger.Logger)
		assert.Equal(t, log.InfoLevel, logger.Logger.Level)
	})

	t.Run("creates logger with text formatter by default", func(t *testing.T) {
		os.Unsetenv("LOG_FORMAT")

		logger := NewLogger()

		require.NotNil(t, logger)
		_, ok := logger.Logger.Formatter.(*log.TextFormatter)
		assert.True(t, ok, "Expected TextFormatter")
	})
}

func TestGetLogLevelByEnvironment(t *testing.T) {
	tests := []struct {
		name          string
		envValue      string
		expectedLevel log.Level
	}{
		{
			name:          "defaults to info when env not set",
			envValue:      "",
			expectedLevel: log.InfoLevel,
		},
		{
			name:          "parses debug level",
			envValue:      "debug",
			expectedLevel: log.DebugLevel,
		},
		{
			name:          "parses info level",
			envValue:      "info",
			expectedLevel: log.InfoLevel,
		},
		{
			name:          "parses warn level",
			envValue:      "warn",
			expectedLevel: log.WarnLevel,
		},
		{
			name:          "parses error level",
			envValue:      "error",
			expectedLevel: log.ErrorLevel,
		},
		{
			name:          "parses fatal level",
			envValue:      "fatal",
			expectedLevel: log.FatalLevel,
		},
		{
			name:          "parses panic level",
			envValue:      "panic",
			expectedLevel: log.PanicLevel,
		},
		{
			name:          "parses trace level",
			envValue:      "trace",
			expectedLevel: log.TraceLevel,
		},
		{
			name:          "handles uppercase level",
			envValue:      "DEBUG",
			expectedLevel: log.DebugLevel,
		},
		{
			name:          "handles mixed case level",
			envValue:      "DeBuG",
			expectedLevel: log.DebugLevel,
		},
		{
			name:          "defaults to info for invalid level",
			envValue:      "invalid",
			expectedLevel: log.InfoLevel,
		},
		{
			name:          "defaults to info for empty string",
			envValue:      "",
			expectedLevel: log.InfoLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				cleanup := testutils.SetEnv(t, "LOG_LEVEL", tt.envValue)
				defer cleanup()
			} else {
				os.Unsetenv("LOG_LEVEL")
			}

			level := getLogLevelByEnvironment()
			assert.Equal(t, tt.expectedLevel, level)
		})
	}
}

func TestNewFormatter(t *testing.T) {
	tests := []struct {
		name             string
		envValue         string
		expectedFormat   string
		validateTemplate func(t *testing.T, formatter log.Formatter)
	}{
		{
			name:           "creates text formatter by default",
			envValue:       "",
			expectedFormat: "text",
			validateTemplate: func(t *testing.T, formatter log.Formatter) {
				textFormatter, ok := formatter.(*log.TextFormatter)
				assert.True(t, ok, "Expected TextFormatter")
				if ok {
					assert.True(t, textFormatter.FullTimestamp)
				}
			},
		},
		{
			name:           "creates text formatter when specified",
			envValue:       "text",
			expectedFormat: "text",
			validateTemplate: func(t *testing.T, formatter log.Formatter) {
				textFormatter, ok := formatter.(*log.TextFormatter)
				assert.True(t, ok, "Expected TextFormatter")
				if ok {
					assert.True(t, textFormatter.FullTimestamp)
				}
			},
		},
		{
			name:           "creates JSON formatter when specified",
			envValue:       "json",
			expectedFormat: "json",
			validateTemplate: func(t *testing.T, formatter log.Formatter) {
				_, ok := formatter.(*log.JSONFormatter)
				assert.True(t, ok, "Expected JSONFormatter")
			},
		},
		{
			name:           "handles uppercase format",
			envValue:       "JSON",
			expectedFormat: "json",
			validateTemplate: func(t *testing.T, formatter log.Formatter) {
				// Should default to text for uppercase (case sensitive check)
				_, ok := formatter.(*log.TextFormatter)
				assert.True(t, ok, "Expected TextFormatter for uppercase")
			},
		},
		{
			name:           "defaults to text for invalid format",
			envValue:       "invalid",
			expectedFormat: "text",
			validateTemplate: func(t *testing.T, formatter log.Formatter) {
				_, ok := formatter.(*log.TextFormatter)
				assert.True(t, ok, "Expected TextFormatter for invalid format")
			},
		},
		{
			name:           "handles format with whitespace",
			envValue:       "  text  ",
			expectedFormat: "text",
			validateTemplate: func(t *testing.T, formatter log.Formatter) {
				_, ok := formatter.(*log.TextFormatter)
				assert.True(t, ok, "Expected TextFormatter with trimmed whitespace")
			},
		},
		{
			name:           "handles JSON format with whitespace",
			envValue:       "  json  ",
			expectedFormat: "json",
			validateTemplate: func(t *testing.T, formatter log.Formatter) {
				_, ok := formatter.(*log.JSONFormatter)
				assert.True(t, ok, "Expected JSONFormatter with trimmed whitespace")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				cleanup := testutils.SetEnv(t, "LOG_FORMAT", tt.envValue)
				defer cleanup()
			} else {
				os.Unsetenv("LOG_FORMAT")
			}

			formatter := newFormatter()
			require.NotNil(t, formatter)
			
			if tt.validateTemplate != nil {
				tt.validateTemplate(t, formatter)
			}
		})
	}
}

func TestNewLogger_WithEnvironmentVariables(t *testing.T) {
	t.Run("respects LOG_LEVEL environment variable", func(t *testing.T) {
		cleanup := testutils.SetEnv(t, "LOG_LEVEL", "debug")
		defer cleanup()

		logger := NewLogger()
		assert.Equal(t, log.DebugLevel, logger.Logger.Level)
	})

	t.Run("respects LOG_FORMAT environment variable", func(t *testing.T) {
		cleanup := testutils.SetEnv(t, "LOG_FORMAT", "json")
		defer cleanup()

		logger := NewLogger()
		_, ok := logger.Logger.Formatter.(*log.JSONFormatter)
		assert.True(t, ok, "Expected JSONFormatter")
	})

	t.Run("respects both LOG_LEVEL and LOG_FORMAT", func(t *testing.T) {
		cleanupLevel := testutils.SetEnv(t, "LOG_LEVEL", "warn")
		defer cleanupLevel()
		
		cleanupFormat := testutils.SetEnv(t, "LOG_FORMAT", "json")
		defer cleanupFormat()

		logger := NewLogger()
		
		assert.Equal(t, log.WarnLevel, logger.Logger.Level)
		_, ok := logger.Logger.Formatter.(*log.JSONFormatter)
		assert.True(t, ok, "Expected JSONFormatter")
	})
}

func TestNewLogger_Integration(t *testing.T) {
	t.Run("logger can log messages", func(t *testing.T) {
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("LOG_FORMAT")

		logger := NewLogger()
		require.NotNil(t, logger)

		// Should not panic when logging
		logger.Info("Test info message")
		logger.Debug("Test debug message") // Won't show at info level
		logger.Warn("Test warn message")
	})

	t.Run("logger with fields", func(t *testing.T) {
		logger := NewLogger()
		require.NotNil(t, logger)

		// Add fields
		loggerWithFields := logger.WithFields(log.Fields{
			"component": "test",
			"action":    "testing",
		})

		// Should not panic
		loggerWithFields.Info("Test message with fields")
	})

	t.Run("debug level logger shows debug messages", func(t *testing.T) {
		cleanup := testutils.SetEnv(t, "LOG_LEVEL", "debug")
		defer cleanup()

		logger := NewLogger()
		assert.Equal(t, log.DebugLevel, logger.Logger.Level)
		
		// Should not panic
		logger.Debug("This debug message should be visible")
	})
}

func TestLogging_ErrorCases(t *testing.T) {
	t.Run("handles very long log level string", func(t *testing.T) {
		cleanup := testutils.SetEnv(t, "LOG_LEVEL", "verylonginvalidloglevelstringthatdoesnotexist")
		defer cleanup()

		level := getLogLevelByEnvironment()
		assert.Equal(t, log.InfoLevel, level, "Should default to info for invalid long string")
	})

	t.Run("handles log level with special characters", func(t *testing.T) {
		cleanup := testutils.SetEnv(t, "LOG_LEVEL", "info#@!")
		defer cleanup()

		level := getLogLevelByEnvironment()
		assert.Equal(t, log.InfoLevel, level, "Should default to info for invalid characters")
	})

	t.Run("handles empty format after trimming", func(t *testing.T) {
		cleanup := testutils.SetEnv(t, "LOG_FORMAT", "    ")
		defer cleanup()

		formatter := newFormatter()
		_, ok := formatter.(*log.TextFormatter)
		assert.True(t, ok, "Should default to text for empty string")
	})
}

func TestLogging_ConcurrentAccess(t *testing.T) {
	t.Run("multiple loggers can be created concurrently", func(t *testing.T) {
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("LOG_FORMAT")

		done := make(chan bool, 10)
		
		for i := 0; i < 10; i++ {
			go func() {
				logger := NewLogger()
				logger.Info("Concurrent log message")
				done <- true
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}
