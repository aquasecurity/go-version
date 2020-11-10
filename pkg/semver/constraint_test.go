package semver

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConstraints(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{">= 1.1", false},
		{">40.50.60, < 50.70", false},
		{"2.0", false},
		{"2.3.5-20161202202307-sha.e8fc5e5", false},
		{">= bar", true},
		{"BAR >= 1.2.3", true},

		// Test with commas separating AND
		{">= 1.2.3, < 2.0", false},
		{">= 1.2.3, < 2.0 || => 3.0, < 4", false},

		// The 3 - 4 should be broken into 2 by the range rewriting
		//{"3 - 4 || => 3.0, < 4", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			_, err := NewConstraints(tt.input)
			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConstraint_Check(t *testing.T) {
	tests := []struct {
		constraint string
		version    string
		want       bool
	}{
		// Equal: =
		{"=2.0.0", "1.2.3", false},
		{"=2.0.0", "2.0.0", true},
		{"=2.0", "1.2.3", false},
		{"=2.0", "2.0.0", true},
		{"=2.0", "2.0.1", true},
		{"=0", "1.0.0", false},

		// Equal: ==
		{"== 2.0.0", "1.2.3", false},
		{"==2.0.0", "2.0.0", true},
		{"== 2.0", "1.2.3", false},
		{"==2.0", "2.0.0", true},
		{"== 2.0", "2.0.1", true},
		{"==0", "1.0.0", false},

		// Equal without "="
		{"4.1", "4.1.0", true},
		{"2", "1.0.0", false},
		{"2", "3.4.5", false},
		{"2", "2.1.1", true},
		{"2.1", "2.1.1", true},
		{"2.1", "2.2.1", false},

		// Not equal
		{"!=4.1.0", "4.1.0", false},
		{"!=4.1.0", "4.1.1", true},
		{"!=4.1", "4.1.0", false},
		{"!=4.1", "4.1.1", false},
		{"!=4.1", "5.1.0-alpha.1", false},
		{"!=4.1-alpha", "4.1.0", false},
		{"!=4.1", "5.1.0", true},

		// Less than
		{"<11", "0.1.0", true},
		{"<11", "11.1.0", false},
		{"<1.1", "0.1.0", true},
		{"<1.1", "1.1.0", false},
		{"<1.1", "1.1.1", false},
		{"<0", "0.0.0-alpha", false},
		{"<0.0.0-z", "0.0.0-alpha", true},

		// Less than or equal
		{"<=11", "1.2.3", true},
		{"<=11", "12.2.3", false},
		{"<=11", "11.2.3", true},
		{"<=1.1", "1.2.3", false},
		{"<=1.1", "0.1.0", true},
		{"<=1.1", "1.1.0", true},
		{"<=1.1", "1.1.1", true},

		// Greater than
		{">1.1", "4.1.0", true},
		{">1.1", "1.1.0", false},
		{">0", "0.0.0", false},
		{">0", "1.0.0", true},
		{">0", "0.0.1-alpha", false},
		{">0.0", "0.0.1-alpha", false},
		{">0", "0.0.0-alpha", false},
		{">0.0.0-0", "0.0.0-alpha", true},
		{">0.0.0-0", "0.0.0-alpha", true},
		{">1.2.3-alpha.1", "1.2.3-alpha.2", true},
		{">1.2.3-alpha.1", "1.3.3-alpha.2", true},
		{">11", "11.1.0", false},
		{">11.1", "11.1.0", false},
		{">11.1", "11.1.1", false},
		{">11.1", "11.2.1", true},

		// Greater than or equal
		{">=11", "11.1.2", true},
		{">=11.1", "11.1.2", true},
		{">=11.1", "11.0.2", false},
		{">=1.1", "4.1.0", true},
		{">=1.1", "1.1.0", true},
		{">=1.1", "0.0.9", false},
		{">=0", "0.0.1-alpha", false},
		{">=0.0", "0.0.1-alpha", false},
		{">=0", "0.0.0-alpha", false},
		{">=0.0.0-0", "0.0.0-alpha", true},
		{">=0.0.0-0", "1.2.3", true},
		{">=0.0.0-0", "3.4.5-beta.1", true},
		{">=0", "0.0.0", true},

		// Asterisk
		{"*", "1.0.0", true},
		{"*", "4.5.6", true},
		{"*", "1.2.3-alpha.1", false},
		{"2.*", "1.0.0", false},
		{"2.*", "3.4.5", false},
		{"2.*", "2.1.1", true},
		{"2.1.*", "2.1.1", true},
		{"2.1.*", "2.2.1", false},

		// Empty: an empty string is treated as * or wild card
		{"", "1.0.0", true},
		{" ", "1.0.0", true},
		{"	", "1.0.0", true},
		{"", "4.5.6", true},
		{"", "1.2.3-alpha.1", false},

		// Tilde
		{"~1.2.3", "1.2.4", true},
		{"~1.2.3", "1.3.4", false},
		{"~1.2", "1.2.4", true},
		{"~1.2", "1.3.4", false},
		{"~1", "1.2.4", true},
		{"~1", "2.3.4", false},
		{"~0.2.3", "0.2.5", true},
		{"~0.2.3", "0.3.5", false},
		{"~1.2.3-beta.2", "1.2.3-beta.4", true},

		// Caret
		// https://docs.npmjs.com/cli/v6/using-npm/semver#caret-ranges-123-025-004
		{"^1.2.3", "1.8.9", true},
		{"^1.2.3", "2.8.9", false},
		{"^1.2.3", "1.2.1", false},
		{"^1.1.0", "2.1.0", false},
		{"^1.2.0", "2.2.1", false},
		{"^1.2.0", "1.2.1-alpha.1", false},
		{"^1.2.0-alpha.2", "1.2.0-alpha.1", false},
		{"^1.2", "1.8.9", true},
		{"^1.2", "2.8.9", false},
		{"^1", "1.8.9", true},
		{"^1", "2.8.9", false},
		{"^0.2.3", "0.2.5", true},
		{"^0.2.3", "0.5.6", false},
		{"^0.2", "0.2.5", true},
		{"^0.2", "0.5.6", false},
		{"^0.0.3", "0.0.3", true},
		{"^0.0.3", "0.0.4", false},
		{"^0.0", "0.0.3", true},
		{"^0.0", "0.1.4", false},
		{"^0.0", "1.0.4", false},
		{"^0", "0.2.3", true},
		{"^0", "1.1.4", false},
		{"^0.2.3-beta.2", "0.2.3-beta.4", true},

		// This next test is a case that is different from npm/js semver handling.
		// Their prereleases are only range scoped to patch releases. This is
		// technically not following semver as docs note. In our case we are
		// following semver.
		{"^1.2.0-alpha.0", "1.2.1-alpha.1", true},
		{"^1.2.0-alpha.0", "1.2.1-alpha.0", true},
		{"^0.2.3-beta.2", "0.2.4-beta.2", true},
		{"^0.2.3-beta.2", "0.3.4-beta.2", false},
		{"^0.2.3-beta.2", "0.2.3-beta.2", true},

		// Not supported: always false
		{">0.0-0", "0.0.1-alpha", false},
		{">0-0", "0.0.1-alpha", false},
		{">0-0", "0.0.0-alpha", false},
		{">=0-0", "0.0.1-alpha", false},
		{">=0.0-0", "0.0.1-alpha", false},
		{">=0-0", "0.0.0-alpha", false},
		{"<0-z", "0.0.0-alpha", false},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%s vs %s", tc.constraint, tc.version), func(t *testing.T) {
			c, err := NewConstraints(tc.constraint)
			require.NoError(t, err)

			v, err := Parse(tc.version)
			require.NoError(t, err)

			got := c.Check(v)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestConstraint_CheckWithPreRelease(t *testing.T) {
	tests := []struct {
		constraint string
		version    string
		want       bool
	}{
		// Equal
		{"=2.0.0", "1.2.3", false},
		{"=2.0.0", "2.0.0", true},
		{"=2.0", "1.2.3", false},
		{"=2.0", "2.0.0", true},
		{"=2.0", "2.0.1", true},
		{"4.1", "4.1.0", true},
		{"== 2.0.0", "1.2.3", false},
		{"==2.0.0", "2.0.0", true},
		{"== 2.0", "2.0.0", true},
		{"==2.0", "2.0.1", true},

		// Not equal
		{"!=4.1.0", "4.1.0", false},
		{"!=4.1.0", "4.1.1", true},
		{"!=4.1", "4.1.0", false},
		{"!=4.1", "4.1.1", false},
		{"!=4.1", "5.1.0", true},

		// Less than
		{"<11", "0.1.0", true},
		{"<11", "11.1.0", false},
		{"<1.1", "0.1.0", true},
		{"<1.1", "1.1.0", false},
		{"<1.1", "1.1.1", false},

		// Less than or equal
		{"<=11", "1.2.3", true},
		{"<=11", "12.2.3", false},
		{"<=11", "11.2.3", true},
		{"<=1.1", "1.2.3", false},
		{"<=1.1", "0.1.0", true},
		{"<=1.1", "1.1.0", true},
		{"<=1.1", "1.1.1", true},

		// Greater than
		{">1.1", "4.1.0", true},
		{">1.1", "1.1.0", false},
		{">0", "0.0.0", false},
		{">0", "1.0.0", true},
		{">11", "11.1.0", false},
		{">11.1", "11.1.0", false},
		{">11.1", "11.1.1", false},
		{">11.1", "11.2.1", true},

		// Greater than or equal
		{">=11", "11.1.2", true},
		{">=11.1", "11.1.2", true},
		{">=11.1", "11.0.2", false},
		{">=1.1", "4.1.0", true},
		{">=1.1", "1.1.0", true},
		{">=1.1", "0.0.9", false},
		{">=0", "0.0.0", true},
		{"=0", "1.0.0", false},

		// Asterisk
		{"*", "1.0.0", true},
		{"*", "4.5.6", true},
		{"2.*", "1.0.0", false},
		{"2.*", "3.4.5", false},
		{"2.*", "2.1.1", true},
		{"2.1.*", "2.1.1", true},
		{"2.1.*", "2.2.1", false},

		// Empty
		{"", "1.0.0", true}, // An empty string is treated as * or wild card
		{"", "4.5.6", true},
		{"2", "1.0.0", false},
		{"2", "3.4.5", false},
		{"2", "2.1.1", true},
		{"2.1", "2.1.1", true},
		{"2.1", "2.2.1", false},

		// Tilde
		{"~1.2.3", "1.2.4", true},
		{"~1.2.3", "1.3.4", false},
		{"~1.2", "1.2.4", true},
		{"~1.2", "1.3.4", false},
		{"~1", "1.2.4", true},
		{"~1", "2.3.4", false},
		{"~0.2.3", "0.2.5", true},
		{"~0.2.3", "0.3.5", false},
		{"~1.2.3-beta.2", "1.2.3-beta.4", true},
		{"^1.2.3", "1.8.9", true},
		{"^1.2.3", "2.8.9", false},
		{"^1.2.3", "1.2.1", false},
		{"^1.1.0", "2.1.0", false},
		{"^1.2.0", "2.2.1", false},
		{"^1.2", "1.8.9", true},
		{"^1.2", "2.8.9", false},
		{"^1", "1.8.9", true},
		{"^1", "2.8.9", false},
		{"^0.2.3", "0.2.5", true},
		{"^0.2.3", "0.5.6", false},
		{"^0.2", "0.2.5", true},
		{"^0.2", "0.5.6", false},
		{"^0.0.3", "0.0.3", true},
		{"^0.0.3", "0.0.4", false},
		{"^0.0", "0.0.3", true},
		{"^0.0", "0.1.4", false},
		{"^0.0", "1.0.4", false},
		{"^0", "0.2.3", true},
		{"^0", "1.1.4", false},

		// pre-release: Not equal
		{"!=4.1", "5.1.0-alpha.1", true},
		{"!=4.1-alpha", "4.1.0", false},

		// pre-release: Greater than
		{">0", "0.0.1-alpha", false},
		{">0.0", "0.0.1-alpha", false},
		{">0-0", "0.0.1-alpha", false},
		{">0.0-0", "0.0.1-alpha", false},
		{">0", "0.0.0-alpha", false},
		{">0-0", "0.0.0-alpha", false},
		{">0.0.0-0", "0.0.0-alpha", true},
		{">1.2.3-alpha.1", "1.2.3-alpha.2", true},
		{">1.2.3-alpha.1", "1.3.3-alpha.2", true},

		// pre-release: Less than
		{"<0", "0.0.0-alpha", true},
		{"<0-z", "0.0.0-alpha", false},

		// pre-release: Greater than or equal
		{">=0", "0.0.1-alpha", true},
		{">=0.0", "0.0.1-alpha", true},
		{">=0-0", "0.0.1-alpha", false},
		{">=0.0-0", "0.0.1-alpha", false},
		{">=0", "0.0.0-alpha", true}, // ">=0.*.*-*"
		{">=0-0", "0.0.0-alpha", false},
		{">=0.0.0", "0.0.0-alpha", false},
		{">=0.0.0-0", "0.0.0-alpha", true},
		{">=0.0.0-0", "1.2.3", true},
		{">=0.0.0-0", "3.4.5-beta.1", true},

		// pre-release: Asterisk
		{"*", "1.2.3-alpha.1", true},

		// pre-release: Empty
		{"", "1.2.3-alpha.1", true},

		// pre-release: Tilde
		{"~1.2.3-beta.2", "1.2.3-beta.4", true},
		{"~1.2.3-beta.2", "1.2.4-beta.2", true},
		{"~1.2.3-beta.2", "1.3.4-beta.2", false},

		// pre-release: Caret
		{"^1.2.0", "1.2.1-alpha.1", true},
		{"^1.2.0-alpha.0", "1.2.1-alpha.1", true},
		{"^1.2.0-alpha.0", "1.2.1-alpha.0", true},
		{"^1.2.0-alpha.2", "1.2.0-alpha.1", false},
		{"^0.2.3-beta.2", "0.2.3-beta.4", true},
		{"^0.2.3-beta.2", "0.2.4-beta.2", true},
		{"^0.2.3-beta.2", "0.3.4-beta.2", false},
		{"^0.2.3-beta.2", "0.2.3-beta.2", true},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%s vs %s", tc.constraint, tc.version), func(t *testing.T) {
			c, err := NewConstraints(tc.constraint, WithPreRelease(true))
			require.NoError(t, err)

			v, err := Parse(tc.version)
			require.NoError(t, err)

			got := c.Check(v)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestConstraint_CheckWithPreReleaseAndZeroPadding(t *testing.T) {
	tests := []struct {
		constraint string
		version    string
		want       bool
	}{
		// Equal
		{"=2.0.0", "1.2.3", false},
		{"=2.0.0", "2.0.0", true},

		// Not equal
		{"!=4.1.0", "4.1.0", false},
		{"!=4.1.0", "4.1.1", true},

		// Less than
		{"<0.0.5", "0.1.0", false},
		{"<1.0.0", "0.1.0", true},

		// Less than or equal
		{"<=0.2.3", "1.2.3", false},
		{"<=1.2.3", "1.2.3", true},

		// Greater than
		{">5.0.0", "4.1.0", false},
		{">4.0.0", "4.1.0", true},

		// Greater than or equal
		{">=11.1.3", "11.1.2", false},
		{">=11.1.2", "11.1.2", true},

		// Asterisk
		{"*", "1.0.0", true},
		{"*", "4.5.6", true},
		{"2.*", "1.0.0", false},
		{"2.*", "3.4.5", false},
		{"2.*", "2.1.1", true},
		{"2.1.*", "2.1.1", true},
		{"2.1.*", "2.2.1", false},

		// Empty
		{"", "1.0.0", true}, // An empty string is treated as * or wild card
		{"", "4.5.6", true},
		{"2", "1.0.0", false},
		{"2", "3.4.5", false},
		{"2", "2.1.1", false},   // different
		{"2.1", "2.1.1", false}, // different
		{"2.1", "2.2.1", false},

		// Tilde
		{"~1.2.3", "1.2.4", true},
		{"~1.2.3", "1.3.4", false},
		{"~1.2", "1.2.4", true},
		{"~1.2", "1.3.4", false},
		{"~1", "1.2.4", true},
		{"~1", "2.3.4", false},
		{"~0.2.3", "0.2.5", true},
		{"~0.2.3", "0.3.5", false},
		{"~1.2.3-beta.2", "1.2.3-beta.4", true},

		// Caret
		{"^1.2.3", "1.8.9", true},
		{"^1.2.3", "2.8.9", false},
		{"^1.2.3", "1.2.1", false},
		{"^1.1.0", "2.1.0", false},
		{"^1.2.0", "2.2.1", false},
		{"^1.2", "1.8.9", true},
		{"^1.2", "2.8.9", false},
		{"^1", "1.8.9", true},
		{"^1", "2.8.9", false},
		{"^0.2.3", "0.2.5", true},
		{"^0.2.3", "0.5.6", false},
		{"^0.2", "0.2.5", true},
		{"^0.2", "0.5.6", false},
		{"^0.0.3", "0.0.3", true},
		{"^0.0.3", "0.0.4", false},
		{"^0.0", "0.0.3", true},
		{"^0.0", "0.1.4", false},
		{"^0.0", "1.0.4", false},
		{"^0", "0.2.3", true},
		{"^0", "1.1.4", false},

		// pre-release: Equal
		{"=4.1", "4.1.0-alpha.1", false},
		{"=4.1-alpha", "4.1.0-alpha", true},
		{"== 4.1", "4.1.0-alpha.1", false},
		{"==4.1-alpha", "4.1.0-alpha", true},

		// pre-release: Not equal
		{"!=4.1", "5.1.0-alpha.1", true},
		{"!=4.1-alpha", "4.1.0", true},

		// pre-release: Greater than
		{">0", "0.0.1-alpha", true},     // different
		{">0.0", "0.0.1-alpha", true},   // different
		{">0-0", "0.0.1-alpha", true},   // different
		{">0.0-0", "0.0.1-alpha", true}, // different
		{">0", "0.0.0-alpha", false},
		{">0-0", "0.0.0-alpha", true}, // different
		{">0.0.0-0", "0.0.0-alpha", true},
		{">1.2.3-alpha.1", "1.2.3-alpha.2", true},
		{">1.2.3-alpha.1", "1.3.3-alpha.2", true},

		// pre-release: Less than
		{"<0", "0.0.0-alpha", true},   // different
		{"<0-z", "0.0.0-alpha", true}, // different
		{"<0", "1.0.0-alpha", false},
		{"<1", "1.0.0-alpha", true}, // different

		// pre-release: Greater than or equal
		{">=0", "0.0.1-alpha", true},
		{">=0.0", "0.0.1-alpha", true},
		{">=0-0", "0.0.1-alpha", true},
		{">=0.0-0", "0.0.1-alpha", true},
		{">=0", "0.0.0-alpha", false},
		{">=0-0", "0.0.0-alpha", true},
		{">=0.0.0-0", "0.0.0-alpha", true},
		{">=0.0.0-0", "1.2.3", true},
		{">=0.0.0-0", "3.4.5-beta.1", true},

		// pre-release: Asterisk
		{"*", "1.2.3-alpha.1", true},

		// pre-release: Empty
		{"", "1.2.3-alpha.1", true},

		// pre-release: Tilde
		{"~1.2.3-beta.2", "1.2.3-beta.4", true},
		{"~1.2.3-beta.2", "1.2.4-beta.2", true},
		{"~1.2.3-beta.2", "1.3.4-beta.2", false},

		// pre-release: Caret
		{"^1.2.0", "1.2.1-alpha.1", true},
		{"^1.2.0-alpha.0", "1.2.1-alpha.1", true},
		{"^1.2.0-alpha.0", "1.2.1-alpha.0", true},
		{"^1.2.0-alpha.2", "1.2.0-alpha.1", false},
		{"^0.2.3-beta.2", "0.2.3-beta.4", true},
		{"^0.2.3-beta.2", "0.2.4-beta.2", true},
		{"^0.2.3-beta.2", "0.3.4-beta.2", false},
		{"^0.2.3-beta.2", "0.2.3-beta.2", true},

		// missing patch: Equal
		{"=2.0", "1.2.3", false},
		{"=2.0", "2.0.0", true},
		{"=2.0", "2.0.1", false}, // different
		{"4.1", "4.1.0", true},
		{"=0", "1.0.0", false},

		// missing patch: Not equal
		{"!=4.1", "4.1.0", false},
		{"!=4.1", "4.1.1", true},
		{"!=4.1", "5.1.0", true},

		// missing minor/patch: Less than
		{"<11", "0.1.0", true},
		{"<11", "11.1.0", false},
		{"<1.1", "0.1.0", true},
		{"<1.1", "1.1.0", false},
		{"<1.1", "1.1.1", false},

		// missing minor/patch: Less than or equal
		{"<=11", "1.2.3", true},
		{"<=11", "12.2.3", false},
		{"<=11", "11.2.3", false}, // different
		{"<=1.1", "1.2.3", false},
		{"<=1.1", "0.1.0", true},
		{"<=1.1", "1.1.0", true},
		{"<=1.1", "1.1.1", false}, // different

		// missing minor/patch: Greater than
		{">1.1", "4.1.0", true},
		{">1.1", "1.1.0", false},
		{">0", "0.0.0", false},
		{">0", "1.0.0", true},
		{">11", "11.1.0", true}, // different
		{">11.1", "11.1.0", false},
		{">11.1", "11.1.1", true}, // different
		{">11.1", "11.2.1", true},

		// missing minor/patch: Greater than or equal
		{">=11", "11.1.2", true},
		{">=11.1", "11.1.2", true},
		{">=11.1", "11.0.2", false},
		{">=1.1", "4.1.0", true},
		{">=1.1", "1.1.0", true},
		{">=1.1", "0.0.9", false},
		{">=0", "0.0.0", true},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%s vs %s", tc.constraint, tc.version), func(t *testing.T) {
			c, err := NewConstraints(tc.constraint, WithPreRelease(true), WithZeroPadding(true))
			require.NoError(t, err)

			v, err := Parse(tc.version)
			require.NoError(t, err)

			got := c.Check(v)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestConstraints_Check(t *testing.T) {
	tests := []struct {
		constraint string
		version    string
		want       bool
	}{
		{"*", "1.2.3", true},
		{"~0.0.0", "1.2.3", false},
		{"0.x.x", "1.2.3", false},
		{"0.0.x", "1.2.3", false},
		{"0.0.0", "1.2.3", false},
		{"*", "1.2.3", true},
		{"^0.0.0", "1.2.3", false},
		{"= 2.0", "1.2.3", false},
		{"= 2.0", "2.0.0", true},
		{"4.1", "4.1.0", true},
		{"4.1.x", "4.1.3", true},
		{"1.x", "1.4.0", true},
		{"!=4.1", "4.1.0", false},
		{"!=4.1-alpha", "4.1.0-alpha", false},
		{"!=4.1-alpha", "4.1.1-alpha", false},
		{"!=4.1-alpha", "4.1.0", false},
		{"!=4.1", "5.1.0", true},
		{"!=4.x", "5.1.0", true},
		{"!=4.x", "4.1.0", false},
		{"!=4.1.x", "4.2.0", true},
		{"!=4.2.x", "4.2.3", false},
		{">1.1", "4.1.0", true},
		{">1.1", "1.1.0", false},
		{"<1.1", "0.1.0", true},
		{"<1.1", "1.1.0", false},
		{"<1.1", "1.1.1", false},
		{"<1.x", "1.1.1", false},
		{"<1.x", "0.1.1", true},
		{"<1.x", "2.0.0", false},
		{"<1.1.x", "1.2.1", false},
		{"<1.1.x", "1.1.500", false},
		{"<1.1.x", "1.0.500", true},
		{"<1.2.x", "1.1.1", true},
		{">=1.1", "4.1.0", true},
		{">=1.1", "4.1.0-beta", false},
		{">=1.1", "1.1.0", true},
		{">=1.1", "0.0.9", false},
		{"<=1.1", "0.1.0", true},
		{"<=1.1", "0.1.0-alpha", false},
		{"<=1.1-a", "0.1.0-alpha", false},
		{"<=1.1", "1.1.0", true},
		{"<=1.x", "1.1.0", true},
		{"<=2.x", "3.0.0", false},
		{"<=1.1", "1.1.1", true},
		{"<=1.1.x", "1.2.500", false},
		{"<=4.5", "3.4.0", true},
		{"<=4.5", "3.7.0", true},
		{"<=4.5", "4.6.3", false},
		{"1.1-3", "4.3.2", false},
		{"^1.1", "1.1.1", true},
		{"^1.1", "4.3.2", false},
		{"^1.x", "1.1.1", true},
		{"^2.x", "1.1.1", false},
		{"^1.x", "2.1.1", false},
		{"^1.x", "1.1.1-beta1", false},
		{"^1.1.2-alpha", "1.2.1-beta1", true},
		{"^1.2.x-alpha", "1.1.1-beta1", false},
		{"~*", "2.1.1", true},
		{"~1", "2.1.1", false},
		{"~1", "1.3.5", true},
		{"~1", "1.4.0", true},
		{"~1.x", "2.1.1", false},
		{"~1.x", "1.3.5", true},
		{"~1.x", "1.4.0", true},
		{"~1.1", "1.1.1", true},
		{"~1.1", "1.1.1-alpha", false},
		{"~1.1-alpha", "1.1.1-beta", false},
		{"~1.1.1-beta", "1.1.1-alpha", false},
		{"~1.1.1-beta", "1.1.1", true},
		{"~1.2.3", "1.2.5", true},
		{"~1.2.3", "1.2.2", false},
		{"~1.2.3", "1.3.2", false},
		{"~1.1", "1.2.3", false},
		{"~1.3", "2.4.5", false},

		// compound
		{">1.1, <2", "1.1.1", false},
		{" >1.1 <2 ", "1.1.1", false},
		{">1.1, <2", "1.2.1", true},
		{">1.1  <2", "1.2.1", true},
		{">1.1, <3", "4.3.2", false},
		{">1.1	<3", "4.3.2", false},
		{">=1.1, <2, !=1.2.3", "1.2.3", false},
		{">=1.1 <2 !=1.2.3", "1.2.3", false},
		{">=1.1, <2, !=1.2.3 || > 3", "4.1.2", true},
		{">=1.1 <2 !=1.2.3 || > 3", "4.1.2", true},
		{">=1.1, <2, !=1.2.3 || > 3", "3.1.2", false},
		{">=1.1, <2, !=1.2.3 || >= 3", "3.0.0", true},
		{">=1.1, <2, !=1.2.3 || > 3", "3.0.0", false},
		{">   1.1, <2", "1.2.1", true},
		{">1.1, <  3", "4.3.2", false},
		{">= 1.1, <     2, !=1.2.3", "1.2.3", false},
		{">= 1.1, <2, !=1.2.3 || > 3", "4.1.2", true},
		{">= 1.1, <2, != 1.2.3 || > 3", "3.1.2", false},
		{">= 1.1, <2, != 1.2.3 || >= 3", "3.0.0", true},
		{">= 1.1, <2, !=1.2.3 || > 3", "3.0.0", false},
		{">= 1.1, <2, !=1.2.3 || > 3", "1.2.3", false},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s vs %s", tt.constraint, tt.version), func(t *testing.T) {
			c, err := NewConstraints(tt.constraint)
			require.NoError(t, err)

			v, err := Parse(tt.version)
			require.NoError(t, err)

			got := c.Check(v)
			assert.Equal(t, tt.want, got)
		})
	}
}
