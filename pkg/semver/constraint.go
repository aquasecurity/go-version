package semver

import (
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/xerrors"
)

const cvRegex string = `v?([0-9|x|X|\*]+)(\.[0-9|x|X|\*]+)?(\.[0-9|x|X|\*]+)?` +
	`(-([0-9A-Za-z\-]+(\.[0-9A-Za-z\-]+)*))?` +
	`(\+([0-9A-Za-z\-]+(\.[0-9A-Za-z\-]+)*))?`

var (
	constraintOperators = map[string]operatorFunc{
		"":   constraintEqual,
		"=":  constraintEqual,
		"!=": constraintNotEqual,
		">":  constraintGreaterThan,
		"<":  constraintLessThan,
		">=": constraintGreaterThanEqual,
		"=>": constraintGreaterThanEqual,
		"<=": constraintLessThanEqual,
		"=<": constraintLessThanEqual,
		"~":  constraintTilde,
		"^":  constraintCaret,
	}
	constraintRegexp *regexp.Regexp
)

type operatorFunc func(v, c Version, conf conf) bool

func init() {
	ops := make([]string, 0, len(constraintOperators))
	for k, v := range constraintOperators {
		ops = append(ops, regexp.QuoteMeta(k))
		constraintOperators[k] = preReleaseCheck(v)
	}

	constraintRegexp = regexp.MustCompile(fmt.Sprintf(
		`^\s*(%s)\s*(%s)\s*$`,
		strings.Join(ops, "|"),
		cvRegex))
}

type Constraints struct {
	constraints [][]constraint
	conf        conf
}

// Constraints is one or more constraint that a semantic version can be
// checked against.
type constraint struct {
	version  Version
	operator operatorFunc
}

func NewConstraints(v string, opts ...ConstraintOption) (Constraints, error) {
	c := new(conf)

	// Apply options
	for _, o := range opts {
		o.apply(c)
	}

	var css [][]constraint
	for _, vv := range strings.Split(v, "||") {
		var cs []constraint
		for _, single := range strings.Split(vv, ",") {
			c, err := newConstraint(single)
			if err != nil {
				return Constraints{}, err
			}
			cs = append(cs, c)
		}
		css = append(css, cs)
	}

	return Constraints{
		constraints: css,
		conf:        *c,
	}, nil

}

func newConstraint(c string) (constraint, error) {
	if c == "" {
		return constraint{
			version:  Version{majorX: true},
			operator: constraintOperators[""],
		}, nil
	}
	m := constraintRegexp.FindStringSubmatch(c)
	if m == nil {
		return constraint{}, xerrors.Errorf("improper constraint: %s", c)
	}

	ver, major, preRelease := m[2], m[3], m[6]
	minor := strings.TrimPrefix(m[4], ".")
	patch := strings.TrimPrefix(m[5], ".")

	var majorX, minorX, patchX bool
	switch {
	case isX(major) || major == "":
		ver = "0.0.0"
		majorX = true
	case isX(minor) || minor == "":
		ver = fmt.Sprintf("%s.0.0%s", major, preRelease)
		minorX = true
	case isX(patch) || patch == "":
		ver = fmt.Sprintf("%s.%s.0%s", major, minor, preRelease)
		patchX = true
	}

	v, err := NewVersion(ver)
	if err != nil {
		return constraint{}, xerrors.Errorf("version parse error (%s): %w", ver, err)
	}

	v.majorX, v.minorX, v.patchX = majorX, minorX, patchX

	return constraint{
		version:  v,
		operator: constraintOperators[m[1]],
	}, nil
}

func (c constraint) check(v Version, conf conf) bool {
	return c.operator(v, c.version, conf)
}

// Check tests if a version satisfies all the constraints.
func (cs Constraints) Check(v Version) bool {
	for _, c := range cs.constraints {
		if andCheck(v, c, cs.conf) {
			return true
		}
	}

	return false
}

func andCheck(v Version, constraints []constraint, conf conf) bool {
	for _, c := range constraints {
		if !c.check(v, conf) {
			return false
		}
	}
	return true
}

func isX(x string) bool {
	switch x {
	case "x", "*", "X":
		return true
	default:
		return false
	}
}

//-------------------------------------------------------------------
// Constraint functions
//-------------------------------------------------------------------

func constraintEqual(v, c Version, conf conf) bool {
	// '*' and 'x' are special cases
	if c.majorX {
		return true
	}
	if !conf.zeroPadding && (c.minorX || c.patchX) {
		// "=1"   => 1.0.0 <= x < 2.0.0
		// "=1.2" => 1.2.0 <= x < 1.3.0
		return constraintTilde(v, c, conf)
	}
	return v.Equal(c)
}

func constraintNotEqual(v, c Version, conf conf) bool {
	// '*' and 'x' are special cases
	if c.majorX {
		return false
	}

	if !conf.zeroPadding {
		if c.minorX {
			return !constraintTilde(v, c, conf)
		} else if c.patchX {
			v.patch = 0
		}
	}
	return !v.Equal(c)
}

func constraintGreaterThan(v, c Version, conf conf) bool {
	if !conf.includePreRelease && (v.preRelease != "" && c.preRelease == "") {
		return false
	}

	if !conf.zeroPadding && (c.majorX || c.minorX || c.patchX) {
		// ">1"   => 2.0.0 <= x
		// ">1.2" => 1.3.0 <= x
		return v.GreaterThanOrEqual(c.TildeBump())
	}
	return v.GreaterThan(c)
}

func constraintLessThan(v, c Version, conf conf) bool {
	if !conf.zeroPadding {
		switch {
		case c.minorX && (v.major == c.major):
			return false
		case c.patchX && (v.major == c.major && v.minor == c.minor):
			return false
		}
	}
	return v.LessThan(c)
}

func constraintGreaterThanEqual(v, c Version, _ conf) bool {
	return v.GreaterThanOrEqual(c)
}

func constraintLessThanEqual(v, c Version, conf conf) bool {
	if !conf.zeroPadding && (c.majorX || c.minorX || c.patchX) {
		// "<=1"   => x < 2.0.0
		// "<=1.2" => x < 1.3.0
		return v.LessThan(c.TildeBump())
	}
	return v.LessThanOrEqual(c)
}

// ~*, ~>* --> >= 0.0.0 (any)
// ~2, ~2.x, ~2.x.x, ~>2, ~>2.x ~>2.x.x --> >=2.0.0, <3.0.0
// ~2.0, ~2.0.x, ~>2.0, ~>2.0.x --> >=2.0.0, <2.1.0
// ~1.2, ~1.2.x, ~>1.2, ~>1.2.x --> >=1.2.0, <1.3.0
// ~1.2.3, ~>1.2.3 --> >=1.2.3, <1.3.0
// ~1.2.0, ~>1.2.0 --> >=1.2.0, <1.3.0
func constraintTilde(v, c Version, _ conf) bool {
	if c.majorX {
		return true
	}
	return v.GreaterThanOrEqual(c) && v.LessThan(c.TildeBump())
}

// ^*      -->  (any)
// ^1.2.3  -->  >=1.2.3 <2.0.0
// ^1.2    -->  >=1.2.0 <2.0.0
// ^1      -->  >=1.0.0 <2.0.0
// ^0.2.3  -->  >=0.2.3 <0.3.0
// ^0.2    -->  >=0.2.0 <0.3.0
// ^0.0.3  -->  >=0.0.3 <0.0.4
// ^0.0    -->  >=0.0.0 <0.1.0
// ^0      -->  >=0.0.0 <1.0.0
func constraintCaret(v, c Version, _ conf) bool {
	if c.majorX {
		return true
	}
	return v.GreaterThanOrEqual(c) && v.LessThan(c.CaretBump())
}

func preReleaseCheck(f operatorFunc) operatorFunc {
	return func(v, c Version, conf conf) bool {
		if !conf.includePreRelease && (v.preRelease != "" && c.preRelease == "") {
			return false
		}
		return f(v, c, conf)
	}
}
