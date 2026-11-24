// Package daemon starts the program as a detached background process and manages its PID file.
package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// childEnv marks the re-executed background process.
const childEnv = "TODOAPP_DAEMON_CHILD"

// IsChild reports whether this process is the detached background process.
func IsChild() bool { return os.Getenv(childEnv) == "1" }

// Start re-executes the current binary detached from the terminal with args (which must not
// contain --daemon). The child's stdout and stderr go to logPath. It returns the child's PID.
func Start(args []string, logPath string) (int, error) {
	exe, err := os.Executable()
	if err != nil {
		return 0, fmt.Errorf("find executable: %w", err)
	}
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return 0, fmt.Errorf("open log file: %w", err)
	}
	defer logFile.Close()

	cmd := exec.Command(exe, args...)
	cmd.Env = append(os.Environ(), childEnv+"=1")
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	detach(cmd)
	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("start background process: %w", err)
	}

	// A child that dies right away (bad flag, port in use) should be reported.
	exited := make(chan error, 1)
	go func() { exited <- cmd.Wait() }()
	select {
	case <-exited:
		return 0, fmt.Errorf("background process exited immediately; see %s", logPath)
	case <-time.After(700 * time.Millisecond):
	}
	return cmd.Process.Pid, nil
}

// WritePID writes this process's PID to path and returns a function that removes it
// (only if it still holds our PID).
func WritePID(path string) (func(), error) {
	pid := os.Getpid()
	if err := os.WriteFile(path, []byte(strconv.Itoa(pid)+"\n"), 0o644); err != nil {
		return nil, fmt.Errorf("write pid file: %w", err)
	}
	return func() {
		b, err := os.ReadFile(path)
		if err == nil && strings.TrimSpace(string(b)) == strconv.Itoa(pid) {
			_ = os.Remove(path)
		}
	}, nil
}
