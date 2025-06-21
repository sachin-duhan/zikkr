package util

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

var log zerolog.Logger

// InitLogger initializes the global logger with the specified configuration
func InitLogger(level string, format string, output string) error {
	// Set up output writer
	var w io.Writer = os.Stdout
	if output != "" {
		file, err := os.OpenFile(output, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}
		if format == "text" {
			w = zerolog.MultiLevelWriter(os.Stdout, file)
		} else {
			w = file
		}
	}

	// Configure logger format
	if format == "text" {
		w = zerolog.ConsoleWriter{
			Out:        w,
			TimeFormat: time.RFC3339,
		}
	}

	// Parse log level
	lvl, err := zerolog.ParseLevel(strings.ToLower(level))
	if err != nil {
		return fmt.Errorf("invalid log level %q: %w", level, err)
	}

	// Create logger
	log = zerolog.New(w).With().Timestamp().Logger().Level(lvl)
	return nil
}

// Logger returns the global logger instance
func Logger() *zerolog.Logger {
	return &log
}

// Debug logs a debug message
func Debug(msg string) {
	log.Debug().Msg(msg)
}

// Info logs an info message
func Info(msg string) {
	log.Info().Msg(msg)
}

// Warn logs a warning message
func Warn(msg string) {
	log.Warn().Msg(msg)
}

// Error logs an error message
func Error(msg string, err error) {
	log.Error().Err(err).Msg(msg)
}

// Fatal logs a fatal message and exits
func Fatal(msg string, err error) {
	log.Fatal().Err(err).Msg(msg)
}

// WithFields returns a context logger with the given fields
func WithFields(fields map[string]interface{}) zerolog.Logger {
	return log.With().Fields(fields).Logger()
}
