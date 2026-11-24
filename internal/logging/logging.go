// Package logging configures zerolog.
package logging

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// New builds the application logger. When logFile is set, logs are also appended to it; the
// returned function closes the file.
func New(level string, asJSON bool, logFile string) (zerolog.Logger, func(), error) {
	noop := func() {}
	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		return zerolog.Logger{}, noop, fmt.Errorf("invalid log level %q: %w", level, err)
	}
	var out io.Writer = os.Stderr
	if !asJSON {
		out = zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.TimeOnly}
	}
	closer := noop
	if logFile != "" {
		f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
		if err != nil {
			return zerolog.Logger{}, noop, fmt.Errorf("open log file: %w", err)
		}
		closer = func() { _ = f.Close() }
		var fileOut io.Writer = f
		if !asJSON {
			fileOut = zerolog.ConsoleWriter{Out: f, TimeFormat: time.DateTime, NoColor: true}
		}
		out = zerolog.MultiLevelWriter(out, fileOut)
	}
	return zerolog.New(out).Level(lvl).With().Timestamp().Logger(), closer, nil
}
