package version

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/aquasecurity/go-version/pkg/part"
	"golang.org/x/xerrors"
)

const cvRegex = `v?([0-9|x|X|\*]+(\.[0-9|x|X|\*]+)*)` +
	`(-([0-9A-Za-z\-]+(\.[0-9A-Za-z\-]+)*))?` +
	`(\+([0-9A-Za-z\-]+(\.[0-9A-Za-z\-]+)*))?`

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
	constraintRangeRegexp *regexp.Regexp
	validConstraintRegexp *regexp.Regexp
)

type operatorFunc func(v, c Version) bool

func init() {
	ops := make([]string, 0, len(constraintOperators))
	for k := range constraintOperators {
		ops = append(ops, regexp.QuoteMeta(k))
	}

	constraintRegexp = regexp.MustCompile(fmt.Sprintf(
		`(%s)\s*(%s)`,
		strings.Join(ops, "|"),
		cvRegex))

	constraintRangeRegexp = regexp.MustCompile(fmt.Sprintf(
		`(%s)\s+-\s+(%s)`,
		cvRegex, cvRegex))

	validConstraintRegexp = regexp.MustCompile(fmt.Sprintf(
		`^\s*(\s*(%s)\s*(%s)\s*\,?)*\s*$`,
		strings.Join(ops, "|"),
		cvRegex))
}

// Constraints is one or more constraint that a version can be checked against.
type Constraints struct {
	constraints [][]Constraint
}

type Constraint struct {
	version      Version
	operator     string
	operatorFunc operatorFunc
	original     string
}

// NewConstraints parses a given constraint and returns a new instance of Constraints
func NewConstraints(v string) (Constraints, error) {
	v = rewriteRange(v)

	var css [][]Constraint
	for _, vv := range strings.Split(v, "||") {
		// Validate the segment
		if !validConstraintRegexp.MatchString(vv) {
			return Constraints{}, xerrors.Errorf("improper constraint: %s", vv)
		}

		ss := constraintRegexp.FindAllString(vv, -1)
		if ss == nil {
			ss = append(ss, strings.TrimSpace(vv))
		}

		var cs []Constraint
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

func rewriteRange(i string) string {
	m := constraintRangeRegexp.FindAllStringSubmatch(i, -1)
	if m == nil {
		return i
	}
	o := i
	for _, v := range m {
		t := fmt.Sprintf(">= %s, <= %s.*", v[1], v[11])
		o = strings.Replace(o, v[0], t, 1)
	}
	return o
}

func newConstraint(c string) (Constraint, error) {
	if c == "" {
		return Constraint{
			version: Version{
				segments: part.NewParts("*"),
			},
			operatorFunc: constraintOperators[""],
		}, nil
	}

	m := constraintRegexp.FindStringSubmatch(c)
	if m == nil {
		return Constraint{}, xerrors.Errorf("improper constraint: %s", c)
	}

	var segments []part.Part
	for _, str := range strings.Split(m[3], ".") {
		segments = append(segments, part.NewPart(str))
	}

	v := Version{
		segments:   segments,
		preRelease: part.NewParts(m[6]),
		original:   c,
	}

	return Constraint{
		version:      v,
		operator:     m[1],
		operatorFunc: constraintOperators[m[1]],
		original:     c,
	}, nil
}

func (c Constraint) check(v Version) bool {
	return c.operatorFunc(v, c.version)
}

func (c Constraint) String() string {
	return c.original
}

func (c Constraint) Version() string {
	return c.version.String()
}

func (c Constraint) Operator() string {
	return c.operator
}

func (cs Constraints) List() [][]Constraint {
	return cs.constraints
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

func andCheck(v Version, constraints []Constraint) bool {
	for _, c := range constraints {
		if !c.check(v) {
			return false
		}
	}
	return true
}

//-------------------------------------------------------------------
// Constraint functions
//-------------------------------------------------------------------

func constraintEqual(v, c Version) bool {
	return v.Equal(c)
}

func constraintNotEqual(v, c Version) bool {
	return !v.Equal(c)
}

func constraintGreaterThan(v, c Version) bool {
	return v.GreaterThan(c)
}

func constraintLessThan(v, c Version) bool {
	return v.LessThan(c)
}

func constraintGreaterThanEqual(v, c Version) bool {
	return v.GreaterThanOrEqual(c)
}

func constraintLessThanEqual(v, c Version) bool {
	return v.LessThanOrEqual(c)
}

func constraintPessimistic(v, c Version) bool {
	return v.GreaterThanOrEqual(c) && v.LessThan(c.PessimisticBump())
}

func constraintTilde(v, c Version) bool {
	// ~*, ~>* --> >= 0.0.0 (any)
	// ~2, ~2.x, ~2.x.x, ~>2, ~>2.x ~>2.x.x --> >=2.0.0, <3.0.0
	// ~2.0, ~2.0.x, ~>2.0, ~>2.0.x --> >=2.0.0, <2.1.0
	// ~1.2, ~1.2.x, ~>1.2, ~>1.2.x --> >=1.2.0, <1.3.0
	// ~1.2.3, ~>1.2.3 --> >=1.2.3, <1.3.0
	// ~1.2.0, ~>1.2.0 --> >=1.2.0, <1.3.0
	return v.GreaterThanOrEqual(c) && v.LessThan(c.TildeBump())
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
	return v.GreaterThanOrEqual(c) && v.LessThan(c.CaretBump())
}
