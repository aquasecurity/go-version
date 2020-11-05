package prerelease

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aquasecurity/go-version/pkg/part"
)

func TestCompare(t *testing.T) {
	type args struct {
		p1 string
		p2 string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "happy path",
			args: args{
				p1: "1",
				p2: "2",
			},
			want: -1,
		},
		{
			name: "empty p1 pre-release",
			args: args{
				p1: "",
				p2: "1",
			},
			want: 1,
		},
		{
			name: "empty p2 pre-release",
			args: args{
				p1: "1",
				p2: "",
			},
			want: -1,
		},
		{
			name: "string",
			args: args{
				p1: "alpha",
				p2: "beta",
			},
			want: -1,
		},
		{
			name: "p1 string and digit",
			args: args{
				p1: "alpha.1",
				p2: "beta",
			},
			want: -1,
		},
		{
			name: "p2 string and digit",
			args: args{
				p1: "beta",
				p2: "alpha.1",
			},
			want: 1,
		},
		{
			name: "dot separated",
			args: args{
				p1: "alpha.10",
				p2: "alpha.2",
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Compare(part.NewParts(tt.args.p1),
				part.NewParts(tt.args.p2)))
		})
	}
}
