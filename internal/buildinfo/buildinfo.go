// Package buildinfo carries version details injected at build time.
package buildinfo

// Version is the application version shown in the UI and --help.
const Version = "1.0.0"

// Info holds values injected via -ldflags -X main.buildTime / main.gitCommit.
type Info struct {
	Version   string `json:"version"`
	BuildTime string `json:"build_time"`
	GitCommit string `json:"git_commit"`
}

// New returns Info for the given injected values.
func New(buildTime, gitCommit string) Info {
	return Info{Version: Version, BuildTime: buildTime, GitCommit: gitCommit}
}

// String is a one-line description.
func (i Info) String() string {
	return "todoapp " + i.Version + " (commit " + i.GitCommit + ", built " + i.BuildTime + ")"
}
