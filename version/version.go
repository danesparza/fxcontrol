// Package version provides build version metadata shared by the CLI and API.
//
// This package should not import any other fxcontrol packages.
package version

import (
	"fmt"

	"github.com/hashicorp/go-version"
)

// Version is the major and minor release version.
var Version = "1.1"

// BuildNumber is the patch component, set at build time and defaulting to "0".
var BuildNumber = "0"

// CommitID is the full Git commit hash, set at build time and empty for local builds.
var CommitID string

// Prerelease is a pre-release marker for the version. If this is "-" (dash)
// then it means that it is a final release. Otherwise, this is a pre-release
// such as "dev" (in development), "beta", "rc1", etc.
var Prerelease = "dev"

// SemVer holds the release version parsed during initialization.
// It validates Version and BuildNumber and does not include Prerelease.
var SemVer *version.Version

func getFormattedVersion() string {
	return fmt.Sprintf("%s.%s", Version, BuildNumber)
}

func init() {
	SemVer = version.Must(version.NewVersion(getFormattedVersion()))
}

// Header is the HTTP response header containing the fxcontrol build version.
const Header = "fxcontrol-service-version"

// String returns the release version with the prerelease suffix unless it is "-".
func String() string {
	if Prerelease != "-" {
		return fmt.Sprintf("%s-%s", getFormattedVersion(), Prerelease)
	}
	return getFormattedVersion()
}
