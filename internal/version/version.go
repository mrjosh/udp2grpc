package version

import (
	"flag"
	"runtime"
	"strings"
)

var (
	version      = "v3.9"
	metadata     = ""
	gitCommit    = ""
	gitTreeState = ""
)

// BuildInfo describes the compile time information.
type BuildInfo struct {
	// Version is the current semver.
	Version string `json:"version,omitempty"`
	// BuildType is the current build type, either nightly or release.
	BuildType string `json:"build_type,omitempty"`
	// Branch is the current git branch.
	Branch string `json:"branch,omitempty"`
	// GitCommit is the git sha1.
	GitCommit string `json:"git_commit,omitempty"`
	// GitTreeState is the state of the git tree.
	GitTreeState string `json:"git_tree_state,omitempty"`
	// GoVersion is the version of the Go compiler used.
	GoVersion string `json:"go_version,omitempty"`
	// CompiledBy is the current compiled_by computer
	CompiledBy string `json:"compiled_by,omitempty"`
	// BuildTime is the build time
	BuildTime string `json:"build_time,omitempty"`
}

// GetVersion returns the semver string of the version
func GetVersion() string {
	if metadata == "" {
		return version
	}
	return version + "+" + metadata
}

// GetUserAgent returns a user agent for user with an HTTP client
func GetUserAgent() string {
	return "UDP2gRPC/" + strings.TrimPrefix(GetVersion(), "v")
}

// Get returns build info
func Get() BuildInfo {
	v := BuildInfo{
		Version:      GetVersion(),
		GitCommit:    gitCommit,
		GitTreeState: gitTreeState,
		GoVersion:    runtime.Version(),
	}

	// HACK(bacongobbler): strip out GoVersion during a test run for consistent test output
	if flag.Lookup("test.v") != nil {
		v.GoVersion = ""
	}
	return v
}
