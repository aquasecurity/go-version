package prerelease

import "github.com/aquasecurity/go-version/pkg/part"

func Compare(p1, p2 string) int {
	switch {
	case p1 == p2:
		return 0
	case p1 == "":
		return 1
	case p2 == "":
		return -1
	}

	return comparePreRelease(p1, p2)
}

func comparePreRelease(v1, v2 string) int {
	// split the prerelease versions by .
	p1 := part.NewParts(v1)
	p2 := part.NewParts(v2)
	return p1.Compare(p2, nil)
}
