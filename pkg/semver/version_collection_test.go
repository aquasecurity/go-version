package semver_test

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aquasecurity/go-version/pkg/semver"
)

func TestCollection(t *testing.T) {
	tests := []struct {
		name     string
		versions []string
		want     []string
	}{
		{
			name: "happy path",
			versions: []string{
				"1.1.1",
				"1.0.0",
				"1.2.0",
				"2.0.0",
				"0.7.1",
			},
			want: []string{
				"0.7.1",
				"1.0.0",
				"1.1.1",
				"1.2.0",
				"2.0.0",
			},
		},
		{
			// ref. 11.4
			// https://semver.org/#spec-item-11
			name: "pre-release",
			versions: []string{
				"1.0.0-alpha.1",
				"1.0.0-rc.1",
				"1.0.0",
				"1.0.0-beta.2",
				"1.0.0-beta.11",
				"1.0.0-alpha",
				"1.0.0-alpha.beta",
				"1.0.0-beta",
			},
			want: []string{
				"1.0.0-alpha",
				"1.0.0-alpha.1",
				"1.0.0-alpha.beta",
				"1.0.0-beta",
				"1.0.0-beta.2",
				"1.0.0-beta.11",
				"1.0.0-rc.1",
				"1.0.0",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			versions := make([]semver.Version, len(tt.versions))
			for i, raw := range tt.versions {
				v, err := semver.Parse(raw)
				require.NoError(t, err)
				versions[i] = v
			}

			sort.Sort(semver.Collection(versions))

			got := make([]string, len(versions))
			for i, v := range versions {
				got[i] = v.String()
			}

			assert.Equal(t, tt.want, got)
		})
	}

}
