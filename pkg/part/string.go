package part

import (
	"strings"
)

type String string

func NewString(s string) String {
	return String(s)
}

func (s String) Compare(other Part) int {
	if other == nil {
		return 1
	} else if s == other {
		return 0
	}

	switch o := other.(type) {
	case Int64:
		return 1
	case Uint64:
		return 1
	case String:
		return strings.Compare(string(s), string(o))
	}
	return 0
}

func (s String) IsNull() bool {
	return s == ""
}
