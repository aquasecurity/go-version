package version

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/aquasecurity/go-version/pkg/part"
	"github.com/aquasecurity/go-version/pkg/prerelease"
	"golang.org/x/xerrors"
)

var (
	constraintOperators = map[string]operatorFunc{
		"":   constraintEqual,
		"=":  constraintEqual,
		"==": constraintEqual,
		"!=": constraintNotEqual,
		">":  constraintGreaterThan,
		"<":  constraintLessThan,
		">=": constraintGreaterThanEqual,
		"=>": constraintGreaterThanEqual,
		"<=": constraintLessThanEqual,
		"=<": constraintLessThanEqual,
		"~>": constraintPessimistic,
		"~":  constraintTilde,
		"^":  constraintCaret,
	}
	constraintRegexp      *regexp.Regexp
	validConstraintRegexp *regexp.Regexp
)

type operatorFunc func(v, c Version) bool

const cvRegex = `v?([0-9|x|X|\*]+(\.[0-9|x|X|\*]+)*)` +
	`(-([0-9]+[0-9A-Za-z\-~]*(\.[0-9A-Za-z\-~]+)*)|(-?([A-Za-z\-~]+[0-9A-Za-z\-~]*(\.[0-9A-Za-z\-~]+)*)))?` +
	`(\+([0-9A-Za-z\-~]+(\.[0-9A-Za-z\-~]+)*))?` +
	`?`

func init() {
	ops := make([]string, 0, len(constraintOperators))
	for k := range constraintOperators {
		ops = append(ops, regexp.QuoteMeta(k))
	}

	constraintRegexp = regexp.MustCompile(fmt.Sprintf(
		`(%s)\s*(%s)`,
		strings.Join(ops, "|"),
		cvRegex))

	validConstraintRegexp = regexp.MustCompile(fmt.Sprintf(
		`^\s*(\s*(%s)\s*(%s)\s*\,?)*\s*$`,
		strings.Join(ops, "|"),
		cvRegex))
}

// Constraints is one or more constraint that a version can be checked against.
type Constraints struct {
	constraints [][]constraint
}

type constraint struct {
	version  Version
	operator operatorFunc
	original string
}

// NewConstraints parses a given constraint and returns a new instance of Constraints
func NewConstraints(v string) (Constraints, error) {
	var css [][]constraint
	for _, vv := range strings.Split(v, "||") {
		// Validate the segment
		if !validConstraintRegexp.MatchString(vv) {
			return Constraints{}, xerrors.Errorf("improper constraint: %s", vv)
		}

		ss := constraintRegexp.FindAllString(vv, -1)
		if ss == nil {
			ss = append(ss, strings.TrimSpace(vv))
		}

		var cs []constraint
		for _, single := range ss {
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
	}, nil

}

func newConstraint(c string) (constraint, error) {
	o := prereleaseCheck
	if c == "" {
		return constraint{
			version:  Version{},
			operator: o,
			original: c,
		}, nil
	}

	m := constraintRegexp.FindStringSubmatch(c)
	if m == nil {
		return constraint{}, xerrors.Errorf("improper constraint: %s", c)
	}

	v, err := newConstraintVersion(m[2:])
	if err != nil {
		return constraint{}, xerrors.Errorf("version parse error (%s): %w", m[2], err)
	}

	if len(v.segments) > 0 {
		o = constraintOperators[m[1]]
	}

	return constraint{
		version:  v,
		operator: o,
		original: c,
	}, nil
}

func newConstraintVersion(matches []string) (Version, error) {
	var segments []part.Uint64
	for _, str := range strings.Split(matches[1], ".") {
		if _, err := part.NewAny(str); err == nil {
			break
		}

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
		preRelease:    part.NewParts(pre),
		original:      matches[0],
	}, nil
}

func (c constraint) check(v Version) bool {
	return c.operator(v, c.version)
}

func (c constraint) String() string {
	return c.original
}

// Check tests if a version satisfies all the constraints.
func (cs Constraints) Check(v Version) bool {
	for _, c := range cs.constraints {
		if andCheck(v, c) {
			return true
		}
	}

	return false
}

// Returns the string format of the constraints
func (cs Constraints) String() string {
	var csStr []string
	for _, orC := range cs.constraints {
		var cstr []string
		for _, andC := range orC {
			cstr = append(cstr, andC.String())
		}
		csStr = append(csStr, strings.Join(cstr, ","))
	}

	return strings.Join(csStr, "||")
}

func andCheck(v Version, constraints []constraint) bool {
	for _, c := range constraints {
		if !c.check(v) {
			return false
		}
	}
	return true
}

func prereleaseCheck(v, c Version) bool {
	if !v.preRelease.IsNull() && c.preRelease.IsNull() {
		return false
	}
	return true
}

//-------------------------------------------------------------------
// Constraint functions
//-------------------------------------------------------------------

func constraintEqual(v, c Version) bool {
	if prerelease.Compare(v.preRelease, c.preRelease) != 0 {
		return false
	}
	return v.GreaterThanOrEqual(c) && v.LessThan(c.NextBump())
}

func constraintNotEqual(v, c Version) bool {
	return !constraintEqual(v, c)
}

func constraintGreaterThan(v, c Version) bool {
	if !prereleaseCheck(v, c) {
		return false
	}
	if !v.preRelease.IsNull() {
		return v.GreaterThan(c)
	}
	return v.GreaterThanOrEqual(c.NextBump())
}

func constraintLessThan(v, c Version) bool {
	return prereleaseCheck(v, c) && v.LessThan(c)
}

func constraintGreaterThanEqual(v, c Version) bool {
	return prereleaseCheck(v, c) && v.GreaterThanOrEqual(c)
}

func constraintLessThanEqual(v, c Version) bool {
	return prereleaseCheck(v, c) && v.LessThan(c.NextBump())
}

func constraintPessimistic(v, c Version) bool {
	return prereleaseCheck(v, c) && v.GreaterThanOrEqual(c) && v.LessThan(c.PessimisticBump())
}

func constraintTilde(v, c Version) bool {
	// ~*, ~>* --> >= 0.0.0 (any)
	// ~2, ~2.x, ~2.x.x, ~>2, ~>2.x ~>2.x.x --> >=2.0.0, <3.0.0
	// ~2.0, ~2.0.x, ~>2.0, ~>2.0.x --> >=2.0.0, <2.1.0
	// ~1.2, ~1.2.x, ~>1.2, ~>1.2.x --> >=1.2.0, <1.3.0
	// ~1.2.3, ~>1.2.3 --> >=1.2.3, <1.3.0
	// ~1.2.0, ~>1.2.0 --> >=1.2.0, <1.3.0
	return prereleaseCheck(v, c) && v.GreaterThanOrEqual(c) && v.LessThan(c.TildeBump())
}

func constraintCaret(v, c Version) bool {
	// ^*      -->  (any)
	// ^1.2.3  -->  >=1.2.3 <2.0.0
	// ^1.2    -->  >=1.2.0 <2.0.0
	// ^1      -->  >=1.0.0 <2.0.0
	// ^0.2.3  -->  >=0.2.3 <0.3.0
	// ^0.2    -->  >=0.2.0 <0.3.0
	// ^0.0.3  -->  >=0.0.3 <0.0.4
	// ^0.0    -->  >=0.0.0 <0.1.0
	// ^0      -->  >=0.0.0 <1.0.0
	return prereleaseCheck(v, c) && v.GreaterThanOrEqual(c) && v.LessThan(c.CaretBump())
}
