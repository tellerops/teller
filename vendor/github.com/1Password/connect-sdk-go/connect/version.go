package connect

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// SDKVersion is the latest Semantic Version of the library
// Do not rename this variable without changing the regex in the Makefile
const SDKVersion = "1.5.3"

const VersionHeaderKey = "1Password-Connect-Version"

// expectMinimumConnectVersion returns an error if the provided minimum version for Connect is lower than the version
// reported in the response from Connect.
func expectMinimumConnectVersion(resp *http.Response, minimumVersion version) error {
	serverVersion, err := getServerVersion(resp)
	if err != nil {
		// Return gracefully if server version cannot be determined reliably
		return nil
	}
	if !serverVersion.IsGreaterOrEqualThan(minimumVersion) {
		return fmt.Errorf("need at least version %s of Connect for this function, detected version %s. Please update your Connect server", minimumVersion, serverVersion)
	}
	return nil
}

func getServerVersion(resp *http.Response) (serverVersion, error) {
	versionHeader := resp.Header.Get(VersionHeaderKey)
	if versionHeader == "" {
		// The last version without the version header was v1.2.0
		return serverVersion{
			version:   version{1, 2, 0},
			orEarlier: true,
		}, nil
	}
	return parseServerVersion(versionHeader)
}

type version struct {
	major int
	minor int
	patch int
}

// serverVersion describes the version reported by the server.
type serverVersion struct {
	version
	// orEarlier is true if the version is derived from the lack of a version header from the server.
	orEarlier bool
}

func (v version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.major, v.minor, v.patch)
}

func (v serverVersion) String() string {
	if v.orEarlier {
		return v.version.String() + " (or earlier)"
	}
	return v.version.String()
}

// IsGreaterOrEqualThan returns true if the lefthand-side version is equal to or or a higher version than the provided
// minimum according to the semantic versioning rules.
func (v version) IsGreaterOrEqualThan(min version) bool {
	if v.major != min.major {
		// Different major version
		return v.major > min.major
	}

	if v.minor != min.minor {
		// Same major, but different minor version
		return v.minor > min.minor
	}

	// Same major and minor version
	return v.patch >= min.patch
}

func parseServerVersion(v string) (serverVersion, error) {
	spl := strings.Split(v, ".")
	if len(spl) != 3 {
		return serverVersion{}, errors.New("wrong length")
	}
	var res [3]int
	for i := range res {
		tmp, err := strconv.Atoi(spl[i])
		if err != nil {
			return serverVersion{}, err
		}
		res[i] = tmp
	}
	return serverVersion{
		version: version{
			major: res[0],
			minor: res[1],
			patch: res[2],
		},
	}, nil
}
