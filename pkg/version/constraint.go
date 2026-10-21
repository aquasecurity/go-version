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

type operatorFunc func(v, c Version, conf conf) bool

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
	constraints [][]Constraint
	conf        conf
}

type Constraint struct {
	version      Version
	operator     string
	operatorFunc operatorFunc
	original     string
}

// NewConstraints parses a given constraint and returns a new instance of Constraints
func NewConstraints(v string, opts ...ConstraintOption) (Constraints, error) {
	config := new(conf)

	// Apply options
	for _, o := range opts {
		o.apply(config)
	}

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
		conf:        *config,
	}, nil

}

func newConstraint(c string) (Constraint, error) {
	m := constraintRegexp.FindStringSubmatch(c)
	if m == nil {
		return Constraint{}, xerrors.Errorf("improper constraint: %s", c)
	}

	v, err := Parse(m[2])
	if err != nil {
		return Constraint{}, xerrors.Errorf("version parse error (%s): %w", m[2], err)
	}

	return Constraint{
		version:      v,
		operator:     m[1],
		operatorFunc: constraintOperators[m[1]],
		original:     c,
	}, nil
}

func (c Constraint) check(v Version, conf conf) bool {
	return c.operatorFunc(v, c.version, conf)
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
		if andCheck(v, c, cs.conf) {
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

func andCheck(v Version, constraints []Constraint, conf conf) bool {
	for _, c := range constraints {
		if !c.check(v, conf) {
			return false
		}
	}
	return true
}

func equal(v, c Version, conf conf) bool {
	return v.compare(c, conf) == 0
}

func greaterThan(v, c Version, conf conf) bool {
	return v.compare(c, conf) > 0
}

func greaterThanOrEqual(v, c Version, conf conf) bool {
	return v.compare(c, conf) >= 0
}

func lessThan(v, c Version, conf conf) bool {
	return v.compare(c, conf) < 0
}

func lessThanOrEqual(v, c Version, conf conf) bool {
	return v.compare(c, conf) <= 0
}

//-------------------------------------------------------------------
// Constraint functions
//-------------------------------------------------------------------

func constraintEqual(v, c Version, conf conf) bool {
	return equal(v, c, conf)
}

func constraintNotEqual(v, c Version, conf conf) bool {
	return !equal(v, c, conf)
}

func constraintGreaterThan(v, c Version, conf conf) bool {
	return greaterThan(v, c, conf)
}

func constraintLessThan(v, c Version, conf conf) bool {
	return lessThan(v, c, conf)
}

func constraintGreaterThanEqual(v, c Version, conf conf) bool {
	return greaterThanOrEqual(v, c, conf)
}

func constraintLessThanEqual(v, c Version, conf conf) bool {
	return lessThanOrEqual(v, c, conf)
}

func constraintPessimistic(v, c Version, conf conf) bool {
	return greaterThanOrEqual(v, c, conf) && lessThan(v, c.PessimisticBump(), conf)
}

func constraintTilde(v, c Version, conf conf) bool {
	// ~*, ~>* --> >= 0.0.0 (any)
	// ~2, ~2.x, ~2.x.x, ~>2, ~>2.x ~>2.x.x --> >=2.0.0, <3.0.0
	// ~2.0, ~2.0.x, ~>2.0, ~>2.0.x --> >=2.0.0, <2.1.0
	// ~1.2, ~1.2.x, ~>1.2, ~>1.2.x --> >=1.2.0, <1.3.0
	// ~1.2.3, ~>1.2.3 --> >=1.2.3, <1.3.0
	// ~1.2.0, ~>1.2.0 --> >=1.2.0, <1.3.0
	return greaterThanOrEqual(v, c, conf) && lessThan(v, c.TildeBump(), conf)
}

func constraintCaret(v, c Version, conf conf) bool {
	// ^*      -->  (any)
	// ^1.2.3  -->  >=1.2.3 <2.0.0
	// ^1.2    -->  >=1.2.0 <2.0.0
	// ^1      -->  >=1.0.0 <2.0.0
	// ^0.2.3  -->  >=0.2.3 <0.3.0
	// ^0.2    -->  >=0.2.0 <0.3.0
	// ^0.0.3  -->  >=0.0.3 <0.0.4
	// ^0.0    -->  >=0.0.0 <0.1.0
	// ^0      -->  >=0.0.0 <1.0.0
	return greaterThanOrEqual(v, c, conf) && lessThan(v, c.CaretBump(), conf)
}
