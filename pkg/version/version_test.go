package version

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewVersion(t *testing.T) {
	tests := []struct {
		version string
		wantErr bool
	}{
		{"", true},
		{"1.2.3", false},
		{"1.0", false},
		{"1", false},
		{"1.2.beta", true},
		{"1.21.beta", true},
		{"foo", true},
		{"1.2-5", false},
		{"1.2-beta.5", false},
		{"\n1.2", true},
		{"1.2.0-x.Y.0+metadata", false},
		{"1.2.0-x.Y.0+metadata-width-hypen", false},
		{"1.2.3-rc1-with-hypen", false},
		{"1.2.3.4", false},
		{"1.2.0.4-x.Y.0+metadata", false},
		{"1.2.0.4-x.Y.0+metadata-width-hypen", false},
		{"1.2.0-X-1.2.0+metadata~dist", false},
		{"1.2.3.4-rc1-with-hypen", false},
		{"1.2.3.4", false},
		{"v1.2.3", false},
		{"foo1.2.3", true},
		{"1.7rc2", false},
		{"v1.7rc2", false},
		{"1.0-", false},
	}
	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			_, err := NewVersion(tt.version)
			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestVersion_Compare(t *testing.T) {
	tests := []struct {
		v1, v2 string
		want   int
	}{
		{"1.2.3", "1.4.5", -1},
		{"1.2-beta", "1.2-beta", 0},
		{"1.2", "1.1.4", 1},
		{"1.2", "1.2-beta", 1},
		{"1.2+foo", "1.2+beta", 0},
		{"v1.2", "v1.2-beta", 1},
		{"v1.2+foo", "v1.2+beta", 0},
		{"v1.2.3.4", "v1.2.3.4", 0},
		{"v1.2.0.0", "v1.2", 0},
		{"v1.2.0.0.1", "v1.2", 1},
		{"v1.2", "v1.2.0.0", 0},
		{"v1.2", "v1.2.0.0.1", -1},
		{"v1.2.0.0", "v1.2.0.0.1", -1},
		{"v1.2.3.0", "v1.2.3.4", -1},
		{"1.7rc2", "1.7rc1", 1},
		{"1.7rc2", "1.7", -1},
		{"1.2.0", "1.2.0-X-1.2.0+metadata~dist", 1},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s vs %s", tt.v1, tt.v2), func(t *testing.T) {
			v1, err := NewVersion(tt.v1)
			require.NoError(t, err)

			v2, err := NewVersion(tt.v2)
			require.NoError(t, err)

			assert.Equal(t, tt.want, v1.Compare(v2))
		})
	}
}
func TestVersion_ComparePreRelease(t *testing.T) {
	tests := []struct {
		v1, v2 string
		want   int
	}{
		{"1.2-beta.2", "1.2-beta.2", 0},
		{"1.2-beta.1", "1.2-beta.2", -1},
		{"1.2-beta.2", "1.2-beta.11", -1},
		{"3.2-alpha.1", "3.2-alpha", 1},
		{"1.2-beta.2", "1.2-beta.1", 1},
		{"1.2-beta.11", "1.2-beta.2", 1},
		{"1.2-beta", "1.2-beta.3", -1},
		{"1.2-alpha", "1.2-beta.3", -1},
		{"1.2-beta", "1.2-alpha.3", 1},
		{"3.0-alpha.3", "3.0-rc.1", -1},
		{"3.0-alpha3", "3.0-rc1", -1},
		{"3.0-alpha.1", "3.0-alpha.beta", -1},
		{"5.4-alpha", "5.4-alpha.beta", -1},
		{"v1.2-beta.2", "v1.2-beta.2", 0},
		{"v1.2-beta.1", "v1.2-beta.2", -1},
		{"v3.2-alpha.1", "v3.2-alpha", 1},
		{"v3.2-rc.1-1-g123", "v3.2-rc.2", 1},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s vs %s", tt.v1, tt.v2), func(t *testing.T) {
			v1, err := NewVersion(tt.v1)
			require.NoError(t, err)

			v2, err := NewVersion(tt.v2)
			require.NoError(t, err)

			assert.Equal(t, tt.want, v1.Compare(v2))
		})
	}
}

func TestVersion_String(t *testing.T) {
	tests := []struct {
		version string
		want    string
	}{
		{"1.2.3", "1.2.3"},
		{"1.2-beta", "1.2-beta"},
		{"1.2.0-x.Y.0", "1.2.0-x.Y.0"},
		{"1.2.0-x.Y.0+metadata", "1.2.0-x.Y.0+metadata"},
		{"1.2.0-metadata-1.2.0+metadata~dist", "1.2.0-metadata-1.2.0+metadata~dist"},
		{"17.03.0-ce", "17.3.0-ce"},
	}
	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			v, err := NewVersion(tt.version)
			require.NoError(t, err)

			assert.Equal(t, tt.want, v.String())
		})
	}
}
func TestVersion_Equal(t *testing.T) {
	tests := []struct {
		v1, v2 string
		want   bool
	}{
		{"1.2.3", "1.4.5", false},
		{"1.2-beta", "1.2-beta", true},
		{"1.2", "1.1.4", false},
		{"1.2", "1.2-beta", false},
		{"1.2+foo", "1.2+beta", true},
		{"v1.2", "v1.2-beta", false},
		{"v1.2+foo", "v1.2+beta", true},
		{"v1.2.3.4", "v1.2.3.4", true},
		{"v1.2.0.0", "v1.2", true},
		{"v1.2.0.0.1", "v1.2", false},
		{"v1.2", "v1.2.0.0", true},
		{"v1.2", "v1.2.0.0.1", false},
		{"v1.2.0.0", "v1.2.0.0.1", false},
		{"v1.2.3.0", "v1.2.3.4", false},
		{"1.7rc2", "1.7rc1", false},
		{"1.7rc2", "1.7", false},
		{"1.2.0", "1.2.0-X-1.2.0+metadata~dist", false},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s vs %s", tt.v1, tt.v2), func(t *testing.T) {
			v1, err := NewVersion(tt.v1)
			require.NoError(t, err)

			v2, err := NewVersion(tt.v2)
			require.NoError(t, err)

			assert.Equal(t, tt.want, v1.Equal(v2))
		})
	}
}
func TestVersion_GreaterThan(t *testing.T) {
	tests := []struct {
		v1, v2 string
		want   bool
	}{
		{"1.2.3", "1.4.5", false},
		{"1.2-beta", "1.2-beta", false},
		{"1.2", "1.1.4", true},
		{"1.2", "1.2-beta", true},
		{"1.2+foo", "1.2+beta", false},
		{"v1.2", "v1.2-beta", true},
		{"v1.2+foo", "v1.2+beta", false},
		{"v1.2.3.4", "v1.2.3.4", false},
		{"v1.2.0.0", "v1.2", false},
		{"v1.2.0.0.1", "v1.2", true},
		{"v1.2", "v1.2.0.0", false},
		{"v1.2", "v1.2.0.0.1", false},
		{"v1.2.0.0", "v1.2.0.0.1", false},
		{"v1.2.3.0", "v1.2.3.4", false},
		{"1.7rc2", "1.7rc1", true},
		{"1.7rc2", "1.7", false},
		{"1.2.0", "1.2.0-X-1.2.0+metadata~dist", true},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s vs %s", tt.v1, tt.v2), func(t *testing.T) {
			v1, err := NewVersion(tt.v1)
			require.NoError(t, err)

			v2, err := NewVersion(tt.v2)
			require.NoError(t, err)

			assert.Equal(t, tt.want, v1.GreaterThan(v2))
		})
	}
}
func TestVersion_LessThan(t *testing.T) {
	tests := []struct {
		v1, v2 string
		want   bool
	}{
		{"1.2.3", "1.4.5", true},
		{"1.2-beta", "1.2-beta", false},
		{"1.2", "1.1.4", false},
		{"1.2", "1.2-beta", false},
		{"1.2+foo", "1.2+beta", false},
		{"v1.2", "v1.2-beta", false},
		{"v1.2+foo", "v1.2+beta", false},
		{"v1.2.3.4", "v1.2.3.4", false},
		{"v1.2.0.0", "v1.2", false},
		{"v1.2.0.0.1", "v1.2", false},
		{"v1.2", "v1.2.0.0", false},
		{"v1.2", "v1.2.0.0.1", true},
		{"v1.2.0.0", "v1.2.0.0.1", true},
		{"v1.2.3.0", "v1.2.3.4", true},
		{"1.7rc2", "1.7rc1", false},
		{"1.7rc2", "1.7", true},
		{"1.2.0", "1.2.0-X-1.2.0+metadata~dist", false},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s vs %s", tt.v1, tt.v2), func(t *testing.T) {
			v1, err := NewVersion(tt.v1)
			require.NoError(t, err)

			v2, err := NewVersion(tt.v2)
			require.NoError(t, err)

			assert.Equal(t, tt.want, v1.LessThan(v2))
		})
	}
}

func TestVersion_GreaterThanOrEqual(t *testing.T) {
	tests := []struct {
		v1, v2 string
		want   bool
	}{
		{"1.2.3", "1.4.5", false},
		{"1.2-beta", "1.2-beta", true},
		{"1.2", "1.1.4", true},
		{"1.2", "1.2-beta", true},
		{"1.2+foo", "1.2+beta", true},
		{"v1.2", "v1.2-beta", true},
		{"v1.2+foo", "v1.2+beta", true},
		{"v1.2.3.4", "v1.2.3.4", true},
		{"v1.2.0.0", "v1.2", true},
		{"v1.2.0.0.1", "v1.2", true},
		{"v1.2", "v1.2.0.0", true},
		{"v1.2", "v1.2.0.0.1", false},
		{"v1.2.0.0", "v1.2.0.0.1", false},
		{"v1.2.3.0", "v1.2.3.4", false},
		{"1.7rc2", "1.7rc1", true},
		{"1.7rc2", "1.7", false},
		{"1.2.0", "1.2.0-X-1.2.0+metadata~dist", true},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s vs %s", tt.v1, tt.v2), func(t *testing.T) {
			v1, err := NewVersion(tt.v1)
			require.NoError(t, err)

			v2, err := NewVersion(tt.v2)
			require.NoError(t, err)

			assert.Equal(t, tt.want, v1.GreaterThanOrEqual(v2))
		})
	}
}
func TestVersion_LessThanOrEqual(t *testing.T) {
	tests := []struct {
		v1, v2 string
		want   bool
	}{
		{"1.2.3", "1.4.5", true},
		{"1.2-beta", "1.2-beta", true},
		{"1.2", "1.1.4", false},
		{"1.2", "1.2-beta", false},
		{"1.2+foo", "1.2+beta", true},
		{"v1.2", "v1.2-beta", false},
		{"v1.2+foo", "v1.2+beta", true},
		{"v1.2.3.4", "v1.2.3.4", true},
		{"v1.2.0.0", "v1.2", true},
		{"v1.2.0.0.1", "v1.2", false},
		{"v1.2", "v1.2.0.0", true},
		{"v1.2", "v1.2.0.0.1", true},
		{"v1.2.0.0", "v1.2.0.0.1", true},
		{"v1.2.3.0", "v1.2.3.4", true},
		{"1.7rc2", "1.7rc1", false},
		{"1.7rc2", "1.7", true},
		{"1.2.0", "1.2.0-X-1.2.0+metadata~dist", false},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s vs %s", tt.v1, tt.v2), func(t *testing.T) {
			v1, err := NewVersion(tt.v1)
			require.NoError(t, err)

			v2, err := NewVersion(tt.v2)
			require.NoError(t, err)

			assert.Equal(t, tt.want, v1.LessThanOrEqual(v2))
		})
	}
}
