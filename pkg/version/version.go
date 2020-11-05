package version

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/aquasecurity/go-version/pkg/part"
	"github.com/aquasecurity/go-version/pkg/prerelease"
	"golang.org/x/xerrors"
)

// The compiled regular expression used to test the validity of a version.
var (
	versionRegex *regexp.Regexp
)

// The raw regular expression string used for testing the validity
// of a version.
const (
	regex = `v?([0-9]+(\.[0-9]+)*?)` +
		`(-([0-9]+[0-9A-Za-z\-~]*(\.[0-9A-Za-z\-~]+)*)|(-?([A-Za-z\-~]+[0-9A-Za-z\-~]*(\.[0-9A-Za-z\-~]+)*)))?` +
		`(\+([0-9A-Za-z\-~]+(\.[0-9A-Za-z\-~]+)*))?` +
		`?`
)

// Version represents a single version.
type Version struct {
	segments      part.Parts
	buildMetadata string
	preRelease    string
	original      string
}

func init() {
	versionRegex = regexp.MustCompile("^" + regex + "$")
}

// NewVersion parses the given version and returns a new
// Version.
func NewVersion(v string) (Version, error) {
	matches := versionRegex.FindStringSubmatch(v)
	if matches == nil {
		return Version{}, xerrors.Errorf("malformed version: %s", v)
	}

	var segments []part.Part
	for _, str := range strings.Split(matches[1], ".") {
		val, err := part.NewUint64(str)
		if err != nil {
			return Version{}, xerrors.Errorf("error parsing version: %w", err)
		}

		segments = append(segments, val)
	}

	pre := matches[7]
	if pre == "" {
		pre = matches[4]
	}

	return Version{
		segments:      segments,
		buildMetadata: matches[10],
		preRelease:    pre,
		original:      v,
	}, nil
}

// Compare compares this version to another version. This
// returns -1, 0, or 1 if this version is smaller, equal,
// or larger than the other version, respectively.
func (v Version) Compare(other Version) int {
	// A quick, efficient equality check
	if v.String() == other.String() {
		return 0
	}

	p1 := v.segments.Normalize()
	p2 := other.segments.Normalize()
	if result := p1.Compare(p2, part.Uint64(0)); result != 0 {
		return result
	}

	return prerelease.Compare(v.preRelease, other.preRelease)
}

// Equal tests if two versions are equal.
func (v Version) Equal(o Version) bool {
	return v.Compare(o) == 0
}

// GreaterThan tests if this version is greater than another version.
func (v Version) GreaterThan(o Version) bool {
	return v.Compare(o) > 0
}

// GreaterThanOrEqual tests if this version is greater than or equal to another version.
func (v Version) GreaterThanOrEqual(o Version) bool {
	return v.Compare(o) >= 0
}

// LessThan tests if this version is less than another version.
func (v Version) LessThan(o Version) bool {
	return v.Compare(o) < 0
}

// LessThanOrEqual tests if this version is less than or equal to another version.
func (v Version) LessThanOrEqual(o Version) bool {
	return v.Compare(o) <= 0
}

// String returns the full version string included pre-release
// and metadata information.
//
// This value is rebuilt according to the parsed segments and other
// information. Therefore, ambiguities in the version string such as
// prefixed zeroes (1.04.0 => 1.4.0), `v` prefix (v1.0.0 => 1.0.0), and
// missing parts (1.0 => 1.0.0) will be made into a canonicalized form
// as shown in the parenthesized examples.
func (v *Version) String() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%d", v.segments[0])
	for _, s := range v.segments[1:len(v.segments)] {
		fmt.Fprintf(&buf, ".%d", s)
	}

	if v.preRelease != "" {
		fmt.Fprintf(&buf, "-%s", v.preRelease)
	}
	if v.buildMetadata != "" {
		fmt.Fprintf(&buf, "+%s", v.buildMetadata)
	}

	return buf.String()
}

// Original returns the original parsed version as-is, including any
// potential whitespace, `v` prefix, etc.
func (v *Version) Original() string {
	return v.original
}
