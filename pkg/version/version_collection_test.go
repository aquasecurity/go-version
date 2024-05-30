package version

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
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
				"1.0",
				"1.2",
				"2",
				"0.7.1",
			},
			want: []string{
				"0.7.1",
				"1.0",
				"1.1.1",
				"1.2",
				"2",
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
			versions := make([]Version, len(tt.versions))
			for i, raw := range tt.versions {
				v, err := Parse(raw)
				require.NoError(t, err)
				versions[i] = v
			}

			sort.Sort(Collection(versions))

			got := make([]string, len(versions))
			for i, v := range versions {
				got[i] = v.String()
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCollection_Constraints(t *testing.T) {
	tests := []struct {
		name        string
		versions    []string
		constraints string
		want        string
	}{
		{
			name:        "pessimistic constraints",
			versions:    []string{"3.0.0", "3.1.0", "3.1.1", "3.1.2", "3.1.3", "3.2.0", "4.0.0"},
			constraints: "~> 3.0",
			want:        "3.2.0",
		},
		{
			name:        "caret constraints",
			versions:    []string{"3.0.0", "3.1.0", "3.1.1", "3.1.2", "3.1.3", "3.2.0", "4.0.0"},
			constraints: "^3.1",
			want:        "3.2.0",
		},
		{
			name:        "tilde constraints",
			versions:    []string{"3.0.0", "3.1.0", "3.1.1", "3.1.2", "3.1.3", "3.2.0", "4.0.0"},
			constraints: "~ 3.1",
			want:        "3.1.3",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraints, err := NewConstraints(tt.constraints)
			require.NoError(t, err)

			expected, err := Parse(tt.want)
			require.NoError(t, err)

			versions := make([]Version, len(tt.versions))
			for i, raw := range tt.versions {
				v, err := Parse(raw)
				require.NoError(t, err)
				versions[i] = v
			}

			sort.Sort(sort.Reverse(Collection(versions)))

			for _, ver := range versions {
				if constraints.Check(ver) {
					assert.Equal(t, expected, ver)
					return
				}
			}

			t.Fatalf("failed to satisfy %s", tt.constraints)
		})
	}
}
