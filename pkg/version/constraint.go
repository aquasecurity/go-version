package version

import (
	"fmt"
	"regexp"
	"strings"

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

func init() {
	ops := make([]string, 0, len(constraintOperators))
	for k := range constraintOperators {
		ops = append(ops, regexp.QuoteMeta(k))
	}

	constraintRegexp = regexp.MustCompile(fmt.Sprintf(
		`(%s)\s*(%s)`,
		strings.Join(ops, "|"),
		regex))

	validConstraintRegexp = regexp.MustCompile(fmt.Sprintf(
		`^\s*(\s*(%s)\s*(%s)\s*\,?)*\s*$`,
		strings.Join(ops, "|"),
		regex))
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

func Test2(){
	fmt.Println("test2")
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
	m := constraintRegexp.FindStringSubmatch(c)
	if m == nil {
		return constraint{}, xerrors.Errorf("improper constraint: %s", c)
	}

	v, err := Parse(m[2])
	if err != nil {
		return constraint{}, xerrors.Errorf("version parse error (%s): %w", m[2], err)
	}

	return constraint{
		version:  v,
		operator: constraintOperators[m[1]],
		original: c,
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
