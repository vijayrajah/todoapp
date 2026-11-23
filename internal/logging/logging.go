// Package logging configures zerolog.
package logging

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// New builds the application logger.
func New(level string, asJSON bool) (zerolog.Logger, error) {
	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		return zerolog.Logger{}, fmt.Errorf("invalid log level %q: %w", level, err)
	}
	var out io.Writer = os.Stderr
	if !asJSON {
		out = zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.TimeOnly}
	}
	return zerolog.New(out).Level(lvl).With().Timestamp().Logger(), nil
}
