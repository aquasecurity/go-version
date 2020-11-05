package semver_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aquasecurity/go-version/pkg/semver"
)

func TestNewVersion(t *testing.T) {
	tests := []struct {
		version string
		err     bool
	}{
		{"1.2.3", false},
		{"1.2.3-alpha.01", true},
		{"1.2.3+test.01", false},
		{"1.2.3-alpha.-1", false},
		{"1.0", true},
		{"1", true},
		{"1.2.beta", true},
		{"foo", true},
		{"1.2-5", true},
		{"1.2-beta.5", true},
		{"\n1.2", true},
		{"1.2.0-x.Y.0+metadata", false},
		{"1.2.0-x.Y.0+metadata-width-hypen", false},
		{"1.2.3-rc1-with-hypen", false},
		{"1.2.3.4", true},
		{"1.2.2147483648", false},
		{"1.2147483648.3", false},
		{"2147483648.3.0", false},
		{"1.0.0-alpha", false},
		{"1.0.0-alpha.1", false},
		{"1.0.0-0.3.7", false},
		{"1.0.0-x.7.z.92", false},
		{"1.0.0-x-y-z.-", false},
		{"1.2.3.4", true},
		{"foo1.2.3", true},
		{"1.7rc2", true},
		{"1.0-", true},
		{"v1.2.3", true},
	}

	for _, tc := range tests {
		t.Run(tc.version, func(t *testing.T) {
			_, err := semver.Parse(tc.version)
			if tc.err {
				require.NotNil(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestVersion_Compare(t *testing.T) {
	cases := []struct {
		v1       string
		v2       string
		expected int
	}{
		{"1.2.3", "1.4.5", -1},
		{"2.2.3", "1.5.1", 1},
		{"2.2.3", "2.2.2", 1},
		{"1.0.0-beta.4", "1.0.0-beta.-2", -1},
		{"1.0.0-beta.-2", "1.0.0-beta.-3", -1},
		{"1.0.0-beta.-3", "1.0.0-beta.5", 1},
		{"1.2.3-beta", "1.2.3-beta", 0},
		{"1.2.3", "1.2.3-beta", 1},
		{"1.2.3+foo", "1.2.3+beta", 0},
		{"1.2.3+foo", "1.2.3+beta", 0},
		{"1.2.0", "1.2.0-X-1.2.0+metadata", 1},
	}

	for _, tt := range cases {
		t.Run(fmt.Sprintf("%s vs %s", tt.v1, tt.v2), func(t *testing.T) {
			v1, err := semver.Parse(tt.v1)
			require.NoError(t, err, tt.v1)

			v2, err := semver.Parse(tt.v2)
			require.NoError(t, err, tt.v2)

			assert.Equal(t, tt.expected, v1.Compare(v2))
		})
	}
}
