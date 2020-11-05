package semver

import (
	"bytes"
	"fmt"
	"math"
	"regexp"

	"github.com/aquasecurity/go-version/pkg/prerelease"

	"github.com/aquasecurity/go-version/pkg/part"

	"golang.org/x/xerrors"
)

var (
	// ErrInvalidSemVer is returned when a given version is invalid
	ErrInvalidSemVer = xerrors.New("invalid semantic version")
)

var versionRegex *regexp.Regexp

// SemVerRegex is the regular expression used to parse a SemVer string.
// See: https://semver.org/#is-there-a-suggested-regular-expression-regex-to-check-a-semver-string
const regex string = `^(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)` +
	`(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))` +
	`?(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`

func init() {
	versionRegex = regexp.MustCompile(regex)
}

// Version represents a semantic version.
type Version struct {
	major, minor, patch    part.Uint64
	preRelease             string
	buildMetadata          string
	majorX, minorX, patchX bool
	original               string
}

// NewVersion parses a given version and returns an instance of Version
func NewVersion(v string) (Version, error) {
	m := versionRegex.FindStringSubmatch(v)
	if m == nil {
		return Version{}, ErrInvalidSemVer
	}

	major, err := part.NewUint64(m[versionRegex.SubexpIndex("major")])
	if err != nil {
		return Version{}, xerrors.Errorf("invalid major version: %w", err)
	}

	minor, err := part.NewUint64(m[versionRegex.SubexpIndex("minor")])
	if err != nil {
		return Version{}, xerrors.Errorf("invalid minor version: %w", err)
	}

	patch, err := part.NewUint64(m[versionRegex.SubexpIndex("patch")])
	if err != nil {
		return Version{}, xerrors.Errorf("invalid minor version: %w", err)
	}

	return Version{
		major:         major,
		minor:         minor,
		patch:         patch,
		preRelease:    m[versionRegex.SubexpIndex("prerelease")],
		buildMetadata: m[versionRegex.SubexpIndex("buildmetadata")],
		original:      v,
	}, nil
}

// String converts a Version object to a string.
func (v Version) String() string {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, "%d.%d.%d", v.major, v.minor, v.patch)
	if v.preRelease != "" {
		fmt.Fprintf(&buf, "-%s", v.preRelease)
	}
	if v.buildMetadata != "" {
		fmt.Fprintf(&buf, "+%s", v.buildMetadata)
	}

	return buf.String()
}

// String converts a Version object to a string.
func (v Version) Original() string {
	return v.original
}

// LessThan tests if one version is less than another one.
func (v Version) LessThan(o Version) bool {
	return v.Compare(o) < 0
}

// LessThanOrEqual tests if this version is less than or equal to another version.
func (v Version) LessThanOrEqual(o Version) bool {
	return v.Compare(o) <= 0
}

// GreaterThan tests if one version is greater than another one.
func (v Version) GreaterThan(o Version) bool {
	return v.Compare(o) > 0
}

// GreaterThanOrEqual tests if this version is greater than or equal to another version.
func (v Version) GreaterThanOrEqual(o Version) bool {
	return v.Compare(o) >= 0
}

// Equal tests if two versions are equal to each other.
// Note, versions can be equal with different metadata since metadata
// is not considered part of the comparable version.
func (v Version) Equal(o Version) bool {
	return v.Compare(o) == 0
}

// Compare compares this version to another one. It returns -1, 0, or 1 if
// the version smaller, equal, or larger than the other version.
//
// Versions are compared by X.Y.Z. Build metadata is ignored. Prerelease is
// lower than the version without a prerelease. Compare always takes into account
// prereleases. If you want to work with ranges using typical range syntaxes that
// skip prereleases if the range is not looking for them use constraints.
func (v Version) Compare(o Version) int {
	// Compare the major, minor, and patch version for differences. If a
	// difference is found return the comparison.
	if result := v.major.Compare(o.major); result != 0 {
		return result
	}
	if result := v.minor.Compare(o.minor); result != 0 {
		return result
	}
	if result := v.patch.Compare(o.patch); result != 0 {
		return result
	}

	// At this point the major, minor, and patch versions are the same.
	return prerelease.Compare(v.preRelease, o.preRelease)
}

// TildeBump returns the right version of tilde ranges
// e.g. ~1.2.3 := >=1.2.3 <1.3.0
// In this case, it returns 1.3.0
// ref.https://docs.npmjs.com/cli/v6/using-npm/semver#caret-ranges-123-025-004
func (v Version) TildeBump() Version {
	switch {
	case v.majorX:
		v.major += 1
	case v.minorX:
		// e.g. 1 => 2.0.0
		v.major += 1
	case v.patchX:
		// e.g. 1.2 => 1.3.0
		v.minor += 1
	default:
		// e.g. 1.2.3 => 1.3.0
		v.minor += 1
		v.patch = 0
	}
	v.preRelease = ""
	v.buildMetadata = ""
	return v
}

// CaretBump returns the right version of caret ranges
// e.g. ^1.2.3 := >=1.2.3 <2.0.0
// In this case, it returns 2.0.0
// ref.https://docs.npmjs.com/cli/v6/using-npm/semver#caret-ranges-123-025-004
func (v Version) CaretBump() Version {
	switch {
	case !v.majorX && v.major != 0:
		// e.g. 1.2.3 => 2.0.0
		v.major += 1
		v.minor = 0
		v.patch = 0
	case v.majorX:
		v.major = math.MaxUint64
	case !v.minorX && v.minor != 0:
		// e.g. 0.2.3 => 0.3.0
		v.minor += 1
		v.patch = 0
	case v.minorX:
		// e.g. 0 => 1.0.0
		v.major += 1
	case !v.patchX && v.patch != 0:
		// e.g. 0.0.3 => 0.0.4
		v.patch += 1
	case v.patchX:
		// e.g. 0.0 => 0.1.0
		v.minor += 1
	}
	v.preRelease = ""
	v.buildMetadata = ""
	return v
}
