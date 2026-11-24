// Package config parses command line flags.
package config

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"

	"github.com/spf13/pflag"

	"todoapp/internal/buildinfo"
)

// ErrExit means the program should exit successfully (help or version was printed).
var ErrExit = errors.New("exit")

// Config is the runtime configuration.
type Config struct {
	Addr     string
	DataDir  string
	ReadOnly bool
	LogLevel string
	LogJSON  bool
	LogFile  string // also write logs to this file ("" = stderr only)
	Daemon   bool   // run detached in the background
	PIDFile  string // write the process ID here ("" = none)
}

// DBPath is the SQLite database file location.
func (c Config) DBPath() string {
	return filepath.Join(c.DataDir, "todoapp.db")
}

// DefaultBackupDir is used for auto backups until a location is set in the UI.
func (c Config) DefaultBackupDir() string {
	return filepath.Join(c.DataDir, "backups")
}

// DaemonLogFile is where a background process logs.
func (c Config) DaemonLogFile() string {
	if c.LogFile != "" {
		return c.LogFile
	}
	return filepath.Join(c.DataDir, "todoapp.log")
}

// Parse reads flags from args. Usage and version output go to out.
func Parse(args []string, info buildinfo.Info, out io.Writer) (Config, error) {
	var cfg Config
	var showVersion bool

	fs := pflag.NewFlagSet("todoapp", pflag.ContinueOnError)
	fs.SetOutput(out)
	fs.StringVar(&cfg.Addr, "addr", "127.0.0.1:8080", "address to listen on (use :8080 to expose on all interfaces)")
	fs.StringVar(&cfg.DataDir, "data-dir", "./data", "directory holding the SQLite database")
	fs.BoolVar(&cfg.ReadOnly, "read-only", false, "reject all writes")
	fs.StringVar(&cfg.LogLevel, "log-level", "info", "log level: debug, info, warn, error")
	fs.BoolVar(&cfg.LogJSON, "log-json", false, "log as JSON instead of console format")
	fs.StringVar(&cfg.LogFile, "log-file", "", "also write logs to this file (daemon default: <data-dir>/todoapp.log)")
	fs.BoolVar(&cfg.Daemon, "daemon", false, "start in the background and return")
	fs.StringVar(&cfg.PIDFile, "pid-file", "", "write the process ID to this file (removed on shutdown)")
	fs.BoolVar(&showVersion, "version", false, "print version and exit")
	fs.Usage = func() {
		fmt.Fprintf(out, "%s\n\nUsage: todoapp [flags]\n\n%s", info, fs.FlagUsages())
	}

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, pflag.ErrHelp) {
			return cfg, ErrExit
		}
		return cfg, err
	}
	if showVersion {
		fmt.Fprintln(out, info)
		return cfg, ErrExit
	}
	return cfg, nil
}
